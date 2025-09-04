package handler

import (
	"fmt"
	"github.com/NiClassic/go-cloud/internal/model"
	"github.com/NiClassic/go-cloud/internal/service"
	"github.com/NiClassic/go-cloud/internal/storage"
	"html/template"
	"net/http"
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
	CreatedAt string
	Size      string
}

func toRows(files []*model.File) []fileRow {
	rows := make([]fileRow, len(files))
	for i, f := range files {
		rows[i] = fileRow{
			Name:      f.Name,
			CreatedAt: f.CreatedAt.Format("02 Jan 06 15:04"),
			Size:      humanReadableSize(f.Size),
		}
	}
	return rows
}

func (p *PersonalFileUploadHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	user := ExtractUserOrRedirect(w, r)

	files, err := p.svc.GetUserFiles(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Render(w, p.tmpl, "personal_files.html", "Your Files | Go-Cloud", map[string]any{"Rows": toRows(files)})
}

func (p *PersonalFileUploadHandler) UploadFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "invalid multipart data: "+err.Error(), http.StatusBadRequest)
		return
	}

	user := ExtractUserOrRedirect(w, r)

	if err := p.svc.StoreFiles(r.Context(), user, reader); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	files, err := p.svc.GetUserFiles(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Render(w, p.tmpl, "file_list.html", "Your Files", map[string]any{"Rows": toRows(files)})
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
