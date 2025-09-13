package handler

import (
	"github.com/NiClassic/go-cloud/config"
	"html/template"
	"net/http"
	"path/filepath"
)

func ParseTemplates() (*template.Template, error) {
	dirs := []string{
		"templates/*.html",
		"templates/*/*.html",
	}

	files := []string{}
	for _, dir := range dirs {
		ff, err := filepath.Glob(dir)
		if err != nil {
			panic(err)
		}
		files = append(files, ff...)
	}

	return template.ParseFiles(files...)
}

func Render(w http.ResponseWriter, tmpl *template.Template, isAuthenticated bool, name, title string, data map[string]any) {
	data["Title"] = title
	data["IsAuthenticated"] = isAuthenticated
	if config.Debug {
		tmplNew, err := ParseTemplates()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl = tmplNew
	}
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "template execution error", http.StatusInternalServerError)
	}
}
