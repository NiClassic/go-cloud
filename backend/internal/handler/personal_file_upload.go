package handler

import (
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type PersonalFileUploadHandler struct {
	tmpl *template.Template
}

func NewPersonalFileUploadHandler(tmpl *template.Template) *PersonalFileUploadHandler {
	return &PersonalFileUploadHandler{tmpl}
}

func (p *PersonalFileUploadHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	Render(w, p.tmpl, "personal_files.html", "Your Files | Go-Cloud", map[string]any{})
}

func (p *PersonalFileUploadHandler) UploadFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	reader, err := r.MultipartReader()
	var fileNames []string
	if err != nil {
		http.Error(w, "invalid multipart data: "+err.Error(), http.StatusBadRequest)
		return
	}

	uploadDir := "/home/nico/Code/go/go-cloud/data/nico"

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break // all parts processed
		}
		if err != nil {
			http.Error(w, "error reading upload: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer part.Close()

		if part.FileName() == "" {
			continue
		}

		safeName := filepath.Base(part.FileName())
		dstPath := filepath.Join(uploadDir, safeName)

		dst, err := os.Create(dstPath)
		if err != nil {
			http.Error(w, "cannot create file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = io.Copy(dst, part)
		fileNames = append(fileNames, safeName)
		dst.Close()
		if err != nil {
			http.Error(w, "error saving file: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	Render(w, p.tmpl, "file_list.html", "Your Files", map[string]any{
		"Files": fileNames,
	})
}
