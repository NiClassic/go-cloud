package handler

import (
	"github.com/NiClassic/go-cloud/config"
	"github.com/NiClassic/go-cloud/internal/logger"
	"github.com/NiClassic/go-cloud/internal/model"
	"github.com/NiClassic/go-cloud/internal/service"
	"net/http"
	"path"
	"strings"
)

type FolderHandler struct {
	*baseHandler
	folderSvc *service.FolderService
	fileSvc   *service.FileService
}

func NewFolderHandler(cfg *config.Config, r *Renderer, folderSvc *service.FolderService, fileSvc *service.FileService) *FolderHandler {
	return &FolderHandler{
		baseHandler: newBaseHandler(cfg, r),
		folderSvc:   folderSvc,
		fileSvc:     fileSvc,
	}
}

func (h *FolderHandler) CreateFolder(w http.ResponseWriter, r *http.Request) {
	logger.Request(r)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.InvalidMethod(r)
		return
	}

	user := ExtractUserOrRedirect(w, r)
	if user == nil {
		return
	}

	if err := r.ParseForm(); err != nil {
		logger.Error("could not parse form: %v", err)
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	folderName := strings.TrimSpace(r.FormValue("name"))
	parentPath := strings.TrimSpace(r.FormValue("path"))

	logger.Debug("Creating folder: name='%s', path='%s'", folderName, parentPath)

	if folderName == "" {
		logger.Error("empty folder name provided")
		http.Error(w, "Folder name is required", http.StatusBadRequest)
		return
	}

	if strings.ContainsAny(folderName, "/\\<>:|?*\"") {
		logger.Error("invalid folder name: %s", folderName)
		http.Error(w, "Invalid folder name", http.StatusBadRequest)
		return
	}

	parentPath = strings.TrimPrefix(parentPath, user.Username)
	parentPath = strings.TrimSuffix(parentPath, "/")
	if parentPath == "" {
		parentPath = "/"
	}

	var parentID int64 = -1
	if parentPath != "/" {
		parentFolder, err := h.folderSvc.GetByPath(r.Context(), user.ID, user.Username, parentPath)
		if err != nil {
			logger.Error("could not get parent folder '%s': %v", parentPath, err)
			parentPath = "/"
		} else {
			parentID = parentFolder.ID
		}
	}

	if parentPath == "/" {
		rootFolder, err := h.folderSvc.GetByPath(r.Context(), user.ID, user.Username, "/")
		if err != nil {
			logger.Error("could not get root folder: %v", err)
			parentID = -1
		} else {
			parentID = rootFolder.ID
		}
	}

	newFolderPath := path.Join(parentPath, folderName)
	if !strings.HasPrefix(newFolderPath, "/") {
		newFolderPath = "/" + newFolderPath
	}

	logger.Debug("Creating folder at path: %s with parent ID: %d", newFolderPath, parentID)

	_, err := h.folderSvc.CreateFolder(r.Context(), user.ID, user.Username, parentID, folderName, newFolderPath)
	if err != nil {
		logger.Error("could not create folder: %v", err)
		http.Error(w, "Failed to create folder", http.StatusInternalServerError)
		return
	}

	var displayFolderID int64
	if parentPath == "/" && parentID == -1 {
		rootFolder, _ := h.folderSvc.GetByPath(r.Context(), user.ID, user.Username, "/")
		if rootFolder != nil {
			displayFolderID = rootFolder.ID
		}
	} else {
		displayFolderID = parentID
	}

	folders, files, err := h.folderSvc.GetFolderContents(r.Context(), user.ID, displayFolderID)
	if err != nil {
		logger.Error("could not get folder contents: %v", err)
		folders = []*model.Folder{}
		files = []*model.File{}
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		h.r.Render(w, true, FileRows, "", map[string]any{
			"Folders": foldersToRows(folders, user.Username),
			"Files":   filesToRows(files),
		})
	} else {
		http.Redirect(w, r, "/files"+parentPath, http.StatusSeeOther)
	}
}

func foldersToRows(folders []*model.Folder, username string) []fileRow {
	rows := make([]fileRow, len(folders))
	for i, f := range folders {
		rows[i] = fileRow{
			Name:      f.Name,
			CreatedAt: f.CreatedAt,
			Size:      -1,
			Id:        f.ID,
			IsDir:     true,
			Path:      strings.TrimPrefix(strings.Trim(f.Path, "/"), username),
		}
	}
	return rows
}

func filesToRows(files []*model.File) []fileRow {
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
