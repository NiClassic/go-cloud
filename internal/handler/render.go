package handler

import (
	"github.com/NiClassic/go-cloud/config"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type Template int

const (
	LoginPage Template = iota
	RegisterPage
	DashboardPage
	PersonalFilePage
	FileRows
	LinkSharePage
	LinkSharePasswordPage
	LinkShareDetailPage
	LinkShareCreationPage
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

func pageToTemplateName(template Template) string {
	switch template {
	case LoginPage:
		return "login.html"
	case RegisterPage:
		return "register.html"
	case DashboardPage:
		return "dashboard.html"
	case PersonalFilePage:
		return "personal_files.html"
	case FileRows:
		return "file_rows"
	case LinkSharePage:
		return "view_links.html"
	case LinkSharePasswordPage:
		return "password_upload_link.html"
	case LinkShareDetailPage:
		return "view_upload_link.html"
	case LinkShareCreationPage:
		return "create_upload_link.html"
	default:
		return "not_found.html"
	}
}

func Render(w http.ResponseWriter, tmpl *template.Template, isAuthenticated bool, template Template, title string, data map[string]any) {
	data["Title"] = title
	data["IsAuthenticated"] = isAuthenticated
	data["Template"] = template
	if config.Debug {
		tmplNew, err := ParseTemplates()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl = tmplNew
	}
	if err := tmpl.ExecuteTemplate(w, pageToTemplateName(template), data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "template execution error", http.StatusInternalServerError)
	}
}
