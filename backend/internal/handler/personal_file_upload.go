package handler

import (
	"github.com/NiClassic/go-cloud/internal/service"
	"github.com/NiClassic/go-cloud/internal/storage"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type PersonalFileUploadHandler struct {
	tmpl *template.Template
	sto  *storage.Storage
	svc  *service.PersonalFileService
}

func NewPersonalFileUploadHandler(tmpl *template.Template, sto *storage.Storage, svc *service.PersonalFileService) *PersonalFileUploadHandler {
	return &PersonalFileUploadHandler{tmpl, sto, svc}
}

func (p *PersonalFileUploadHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	user := ExtractUserOrRedirect(w, r)

	files, err := p.svc.GetUserFiles(user.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	Render(w, p.tmpl, "personal_files.html", "Your Files | Go-Cloud", map[string]any{"Files": files})
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
	err = p.sto.CreateBaseDirIfAbsent(user.Username)
	if err != nil {
		http.Error(w, "failed to create base dir: "+err.Error(), http.StatusInternalServerError)
		return
	}

	baseUserPath := p.sto.GetBaseDirForUser(user.Username)

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break // all parts processed
		}
		if err != nil {
			http.Error(w, "error reading upload: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if part.FileName() == "" {
			continue
		}

		safeName := filepath.Base(part.FileName())
		dstPath := filepath.Join(baseUserPath, safeName)

		dst, err := os.Create(dstPath)
		if err != nil {
			http.Error(w, "cannot create file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = io.Copy(dst, part)
		if err != nil {
			http.Error(w, "error saving file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err = part.Close(); err != nil {
			http.Error(w, "cannot close file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err = dst.Close(); err != nil {
			http.Error(w, "cannot close file: "+err.Error(), http.StatusInternalServerError)
			return
		}

	}
	files, err := p.svc.GetUserFiles(user.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	Render(w, p.tmpl, "file_list.html", "Your Files", map[string]any{
		"Files": files,
	})
}
