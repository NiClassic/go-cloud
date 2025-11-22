package handler

import (
	"errors"
	"github.com/NiClassic/go-cloud/config"
	"github.com/NiClassic/go-cloud/internal/logger"
	"github.com/NiClassic/go-cloud/internal/model"
	"github.com/NiClassic/go-cloud/internal/path"
	"github.com/NiClassic/go-cloud/internal/service"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type fileRow struct {
	Name      string
	CreatedAt time.Time
	Size      int64
	Id        int64
	IsDir     bool
	Path      string
}
type FileHandler struct {
	*baseHandler
	fileService   *service.FileService
	folderService *service.FolderService
	shareService  *service.FileShareService
	converter     *path.Converter
}

func NewFileHandler(cfg *config.Config, r *Renderer, fileService *service.FileService, folderService *service.FolderService, shareService *service.FileShareService, c *path.Converter) *FileHandler {
	return &FileHandler{newBaseHandler(cfg, r), fileService, folderService, shareService, c}
}

func (h *FileHandler) filesToRows(files []*model.File) []fileRow {
	rows := make([]fileRow, len(files))
	for i, f := range files {
		rows[i] = fileRow{
			Name:      f.Name,
			CreatedAt: f.CreatedAt,
			Size:      f.Size,
			Id:        f.ID,
			IsDir:     false,
		}
	}
	return rows
}

func (h *FileHandler) foldersToRows(folders []*model.Folder) []fileRow {
	rows := make([]fileRow, len(folders))
	for i, f := range folders {
		rows[i] = fileRow{
			Name:      f.Name,
			CreatedAt: f.CreatedAt,
			Size:      -1,
			Id:        f.ID,
			IsDir:     true,
			Path:      h.converter.ToURLPath(f.Path),
		}
	}
	return rows
}

func (h *FileHandler) RedirectNoTrailingSlash(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/files/", http.StatusSeeOther)
}

func (h *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	logger.Request(r)
	user := ExtractUserOrRedirect(w, r)
	if user == nil {
		return
	}

	dbPath := h.converter.FromURLPath(r.URL.Path)

	folder, err := h.folderService.GetByPath(r.Context(), user.ID, user.Username, dbPath)
	if err != nil {
		logger.Error("could not get folder: %v", err)
		http.Redirect(w, r, "/files/", http.StatusSeeOther)
		return
	}

	folders, files, err := h.folderService.GetFolderContents(r.Context(), user.ID, folder.ID)
	if err != nil {
		logger.Error("could not get folder content: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	breadcrumbs := h.converter.GetBreadcrumbs(dbPath)

	h.r.Render(w, true, PersonalFilePage, "Your Files", map[string]any{
		"Files":               h.filesToRows(files),
		"Folders":             h.foldersToRows(folders),
		"CurrentFolderID":     folder.ID,
		"CurrentFolderPath":   dbPath,
		"CurrentFolderName":   folder.Name,
		"Breadcrumbs":         breadcrumbs,
		"LastBreadcrumbIndex": len(breadcrumbs) - 1,
	})
}

func (h *FileHandler) UploadFiles(w http.ResponseWriter, r *http.Request) {
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

	dbPath := h.converter.FromURLPath(strings.TrimPrefix(r.URL.Path, "/files/upload"))

	folder, err := h.folderService.GetByPath(r.Context(), user.ID, user.Username, dbPath)
	if err != nil {
		logger.Error("could not get folder: %v", err)
		http.Error(w, "folder not found", http.StatusNotFound)
		return
	}

	if err := h.fileService.StoreFiles(r.Context(), user, reader, folder.ID, dbPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("could not store files: %v", err)
		return
	}

	folders, files, err := h.folderService.GetFolderContents(r.Context(), user.ID, folder.ID)
	if err != nil {
		logger.Error("could not get folder content: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.r.Render(w, true, FileRows, "", map[string]any{
		"Files":   h.filesToRows(files),
		"Folders": h.foldersToRows(folders),
	})
}

func (h *FileHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	user := ExtractUserOrRedirect(w, r)
	if user == nil {
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/download/")
	fileID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var (
		rdr  io.ReadCloser
		size int64
		name string
	)
	rdr, size, name, err = h.fileService.DownloadOwn(r.Context(), user.Username, user.ID, fileID)
	if errors.Is(err, service.ErrNotOwner) {
		rdr, size, name, err = h.fileService.DownloadShared(r.Context(), user.ID, fileID)
	}
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer rdr.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	_, _ = io.Copy(w, rdr)
}
