package handler

import (
	"fmt"
	"github.com/NiClassic/go-cloud/config"
	"github.com/NiClassic/go-cloud/internal/logger"
	"github.com/NiClassic/go-cloud/internal/model"
	"github.com/NiClassic/go-cloud/internal/path"
	"github.com/NiClassic/go-cloud/internal/service"
	"github.com/NiClassic/go-cloud/internal/storage"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type fileRow struct {
	Name      string
	CreatedAt time.Time
	Size      string
	Id        int64
	IsDir     bool
	Path      string
}
type PersonalFileUploadHandler struct {
	*baseHandler
	sto           storage.FileManager
	fileService   *service.PersonalFileService
	folderService *service.FolderService
	converter     *path.Converter
}

func NewPersonalFileUploadHandler(cfg *config.Config, r *Renderer, sto storage.FileManager, fileService *service.PersonalFileService, folderService *service.FolderService, c *path.Converter) *PersonalFileUploadHandler {
	return &PersonalFileUploadHandler{newBaseHandler(cfg, r), sto, fileService, folderService, c}
}

func (p *PersonalFileUploadHandler) filesToRows(files []*model.File) []fileRow {
	rows := make([]fileRow, len(files))
	for i, f := range files {
		rows[i] = fileRow{
			Name:      f.Name,
			CreatedAt: f.CreatedAt,
			Size:      p.humanReadableSize(f.Size),
			Id:        f.ID,
			IsDir:     false,
		}
	}
	return rows
}

func (p *PersonalFileUploadHandler) foldersToRows(folders []*model.Folder) []fileRow {
	rows := make([]fileRow, len(folders))
	for i, f := range folders {
		rows[i] = fileRow{
			Name:      f.Name,
			CreatedAt: f.CreatedAt,
			Size:      "—",
			Id:        f.ID,
			IsDir:     true,
			Path:      p.converter.ToURLPath(f.Path),
		}
	}
	return rows
}

func (p *PersonalFileUploadHandler) RedirectNoTrailingSlash(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/files/", http.StatusSeeOther)
}

type breadCrumbItem struct {
	Name    string
	Path    string
	Current bool
}

func (p *PersonalFileUploadHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	logger.Request(r)
	user := ExtractUserOrRedirect(w, r)
	if user == nil {
		return
	}

	// Extract path from URL
	urlPath := r.URL.Path
	dbPath := p.converter.FromURLPath(urlPath)

	// Get folder from database
	folder, err := p.folderService.GetByPath(r.Context(), user.ID, user.Username, dbPath)
	if err != nil {
		logger.Error("could not get folder: %v", err)
		http.Redirect(w, r, "/files/", http.StatusSeeOther)
		return
	}

	// Get folder contents
	folders, files, err := p.folderService.GetFolderContents(r.Context(), user.ID, folder.ID)
	if err != nil {
		logger.Error("could not get folder content: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Generate breadcrumbs
	breadcrumbs := p.converter.GetBreadcrumbs(dbPath)

	p.r.Render(w, true, PersonalFilePage, "Your Files", map[string]any{
		"Files":               p.filesToRows(files),
		"Folders":             p.foldersToRows(folders),
		"CurrentFolderID":     folder.ID,
		"CurrentFolderPath":   dbPath, // Store as DB path
		"CurrentFolderName":   folder.Name,
		"Breadcrumbs":         breadcrumbs,
		"LastBreadcrumbIndex": len(breadcrumbs) - 1,
	})
}

func (p *PersonalFileUploadHandler) UploadFiles(w http.ResponseWriter, r *http.Request) {
	logger.Request(r)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.InvalidMethod(r)
		return
	}

	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "invalid multipart data: "+err.Error(), http.StatusBadRequest)
		logger.Error("invalid multipart data: %v", err)
		return
	}

	user := ExtractUserOrRedirect(w, r)
	if user == nil {
		return
	}

	// Extract path from URL and convert to DB format
	urlPath := strings.TrimPrefix(r.URL.Path, "/files/upload")
	dbPath := p.converter.FromURLPath(urlPath)

	// Get folder from database
	folder, err := p.folderService.GetByPath(r.Context(), user.ID, user.Username, dbPath)
	if err != nil {
		logger.Error("could not get folder: %v", err)
		http.Error(w, "folder not found", http.StatusNotFound)
		return
	}

	// Store files (pass DB path format)
	if err := p.fileService.StoreFiles(r.Context(), user, reader, folder.ID, dbPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("could not store files: %v", err)
		return
	}

	// Get updated folder contents
	folders, files, err := p.folderService.GetFolderContents(r.Context(), user.ID, folder.ID)
	if err != nil {
		logger.Error("could not get folder content: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	p.r.Render(w, true, FileRows, "", map[string]any{
		"Files":   p.filesToRows(files),
		"Folders": p.foldersToRows(folders),
	})
}

func (p *PersonalFileUploadHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	logger.Request(r)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.InvalidMethod(r)
		return
	}

	user := ExtractUserOrRedirect(w, r)
	if user == nil {
		return
	}

	// Extract file ID from URL
	idStr := strings.TrimPrefix(r.URL.Path, "/download/")
	fileID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	// Get file from database
	file, err := p.fileService.GetFileById(r.Context(), fileID)
	if err != nil || file.UserID != user.ID {
		http.NotFound(w, r)
		return
	}

	// File location is already the full filesystem path
	// But we can use pathUtil to verify or rebuild if needed
	fullPath := p.converter.GetFullFilePath(user.Username, file.Location)

	http.ServeFile(w, r, fullPath)
}

func (p *PersonalFileUploadHandler) humanReadableSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
