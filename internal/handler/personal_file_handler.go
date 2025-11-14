package handler

import (
	"fmt"
	"github.com/NiClassic/go-cloud/config"
	"github.com/NiClassic/go-cloud/internal/logger"
	"github.com/NiClassic/go-cloud/internal/model"
	"github.com/NiClassic/go-cloud/internal/service"
	"github.com/NiClassic/go-cloud/internal/storage"
	"net/http"
	"os"
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
}

func NewPersonalFileUploadHandler(cfg *config.Config, r *Renderer, sto storage.FileManager, fileService *service.PersonalFileService, folderService *service.FolderService) *PersonalFileUploadHandler {
	return &PersonalFileUploadHandler{newBaseHandler(cfg, r), sto, fileService, folderService}
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

func (p *PersonalFileUploadHandler) foldersToRows(folders []*model.Folder, username string) []fileRow {
	rows := make([]fileRow, len(folders))
	for i, f := range folders {
		rows[i] = fileRow{
			Name:      f.Name,
			CreatedAt: f.CreatedAt,
			Size:      "—",
			Id:        f.ID,
			IsDir:     true,
			Path:      strings.TrimPrefix(strings.Trim(f.Path, "/"), username),
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

	folderPath := strings.TrimPrefix(r.URL.Path, "/files")
	folderPath = strings.TrimSuffix(folderPath, "/")
	if folderPath == "" {
		folderPath = "/"
	}

	folder, err := p.folderService.GetByPath(r.Context(), user.ID, user.Username, folderPath)
	if err != nil {
		logger.Error("could not get folder: %v", err)
		http.Redirect(w, r, "/files/", http.StatusSeeOther)
		return
	}

	folders, files, err := p.folderService.GetFolderContents(r.Context(), user.ID, folder.ID)
	if err != nil {
		logger.Error("could not get folder content: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	parts := strings.Split(strings.Trim(strings.TrimPrefix(folder.Path, user.Username), "/"), "/")
	var m []breadCrumbItem
	if len(parts) == 1 && parts[0] == "" {
		m = []breadCrumbItem{{
			Name:    "Home",
			Path:    "/",
			Current: true,
		}}
	} else {
		m = make([]breadCrumbItem, len(parts)+1)
		m[0] = breadCrumbItem{
			Name:    "Home",
			Path:    "/",
			Current: false,
		}
		cur := "/"
		for i, p := range parts {
			cur = cur + p + "/"
			m[i+1] = breadCrumbItem{
				Name:    p,
				Path:    cur,
				Current: i == len(parts)-1,
			}
		}
	}

	p.r.Render(w, true, PersonalFilePage, "Your Files", map[string]any{
		"Files":               p.filesToRows(files),
		"Folders":             p.foldersToRows(folders, user.Username),
		"CurrentFolderID":     folder.ID,
		"CurrentFolderPath":   folder.Path,
		"CurrentFolderName":   folder.Name,
		"Breadcrumbs":         m,
		"LastBreadcrumbIndex": len(m) - 1,
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

	folderPath := strings.TrimPrefix(r.URL.Path, "/files/upload/")
	folderPath = strings.TrimPrefix(folderPath, user.Username)
	folderPath = strings.TrimSuffix(folderPath, "/")
	if folderPath == "" {
		folderPath = "/"
	}

	folder, err := p.folderService.GetByPath(r.Context(), user.ID, user.Username, folderPath)
	if err != nil {
		logger.Error("could not get folder: %v", err)
		http.Error(w, "folder not found", http.StatusNotFound)
		return
	}

	if err := p.fileService.StoreFiles(r.Context(), user, reader, folder.ID, folderPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("could not store files: %v", err)
		return
	}

	folders, files, err := p.folderService.GetFolderContents(r.Context(), user.ID, folder.ID)
	if err != nil {
		logger.Error("could not get folder content: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	p.r.Render(w, true, FileRows, "", map[string]any{
		"Files":   p.filesToRows(files),
		"Folders": p.foldersToRows(folders, user.Username),
	})
}

func (p *PersonalFileUploadHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	logger.Request(r)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.InvalidMethod(r)
		return
	}

	suffix := strings.TrimPrefix(r.URL.Path, "/download/")
	fileId, err := strconv.ParseInt(suffix, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		logger.Error("could not parse file id: %s", r.URL.Path)
		return
	}

	file, err := p.fileService.GetFileById(r.Context(), fileId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("could not get file by id: %v", err)
		return
	}

	user := ExtractUserOrRedirect(w, r)
	if user == nil {
		return
	}

	if file.UserID != user.ID {
		http.NotFound(w, r)
		logger.Error("user tried to access file that does not belong to them: %v", r.URL.Path)
		return
	}

	f, err := os.Open(file.Location)
	if err != nil {
		http.Error(w, "cannot open file", http.StatusInternalServerError)
		logger.Error("could not open file: %v", err)
		return
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			http.Error(w, "cannot close file", http.StatusInternalServerError)
			logger.Error("could not close file: %v", err)
			return
		}
	}(f)

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", file.Name))
	w.Header().Set("Content-Type", file.MimeType)
	w.Header().Set("Content-Length", strconv.FormatInt(file.Size, 10))
	w.Header().Set("Cache-Control", "no-store")
	http.ServeContent(w, r, file.Name, file.CreatedAt, f)
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
