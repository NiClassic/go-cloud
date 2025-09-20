package handler

import (
	"fmt"
	"github.com/NiClassic/go-cloud/internal/logger"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/NiClassic/go-cloud/internal/model"
	"github.com/NiClassic/go-cloud/internal/service"
	"github.com/NiClassic/go-cloud/internal/storage"
)

type PersonalFileUploadHandler struct {
	tmpl *template.Template
	sto  *storage.Storage
	svc  *service.PersonalFileService
}

func NewPersonalFileUploadHandler(tmpl *template.Template, sto *storage.Storage, svc *service.PersonalFileService) *PersonalFileUploadHandler {
	return &PersonalFileUploadHandler{tmpl, sto, svc}
}

type fileRow struct {
	Name      string
	CreatedAt time.Time
	Size      string
	Id        int64
}

func toRows(files []*model.File) []fileRow {
	rows := make([]fileRow, len(files))
	for i, f := range files {
		rows[i] = fileRow{
			Name:      f.Name,
			CreatedAt: f.CreatedAt,
			Size:      humanReadableSize(f.Size),
			Id:        f.ID,
		}
	}
	return rows
}

func (p *PersonalFileUploadHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	logger.Request(r)
	user := ExtractUserOrRedirect(w, r)

	files, err := p.svc.GetUserFiles(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("could not get user files: %v", err)
		return
	}

	Render(w, p.tmpl, true, PersonalFilePage, "Your Files", map[string]any{
		"Rows": toRows(files),
		"Now":  time.Now(),
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

	if err := p.svc.StoreFiles(r.Context(), user, reader); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("could not store files: %v", err)
		return
	}

	files, err := p.svc.GetUserFiles(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("could not get user files: %v", err)
		return
	}

	Render(w, p.tmpl, true, FileRows, "Your Files", map[string]any{
		"Rows": toRows(files),
		"Now":  time.Now(),
	})
}

func (p *PersonalFileUploadHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	logger.Request(r)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.InvalidMethod(r)
		return
	}

	suffix := strings.TrimPrefix(r.URL.Path, "/files/")
	parts := strings.SplitN(suffix, "/", 2)
	if len(parts) != 2 || parts[1] != "download" {
		http.NotFound(w, r)
		logger.Error("invalid file path: %s", r.URL.Path)
		return
	}
	fileIdStr := parts[0]
	fileId, err := strconv.ParseInt(fileIdStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		logger.Error("could not parse file id: %s", r.URL.Path)
		return
	}

	file, err := p.svc.GetFileById(r.Context(), fileId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("could not get file by id: %v", err)
		return
	}

	user := ExtractUserOrRedirect(w, r)
	if file.UserID != user.ID {
		http.NotFound(w, r)
		logger.Error("user tried to access file that does not belong to him: %v", r.URL.Path)
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
	w.Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=%q", file.Name))
	w.Header().Set("Content-Type", file.MimeType)
	w.Header().Set("Content-Length", strconv.FormatInt(file.Size, 10))
	w.Header().Set("Cache-Control", "no-store")
	http.ServeContent(w, r, file.Name, file.CreatedAt, f)
}

func humanReadableSize(b int64) string {
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
