package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/NiClassic/go-cloud/config"
	"github.com/NiClassic/go-cloud/internal/logger"
)

type Renderer struct {
	cfg  *config.Config
	tmpl *template.Template
}

func NewRenderer(cfg *config.Config) (*Renderer, error) {
	r := &Renderer{cfg: cfg}
	err := r.parseTemplates()
	return r, err
}

type Template int

const (
	LoginPage Template = iota
	RegisterPage
	DashboardPage
	PersonalFilePage
	SharePage
	FileRows
	LinkSharePage
	LinkSharePasswordPage
	LinkShareDetailPage
	LinkShareCreationPage
)

func (r *Renderer) parseTemplates() error {
	dirs := []string{
		"templates/*.html",
		"templates/*/*.html",
	}

	files := []string{}
	for _, dir := range dirs {
		ff, err := filepath.Glob(dir)
		if err != nil {
			return err
		}
		files = append(files, ff...)
	}

	tmpl := template.New("").Funcs(GetTemplateFunctions())
	finalTemplates, err := tmpl.ParseFiles(files...)
	if err != nil {
		return err
	}
	r.tmpl = finalTemplates
	return nil
}

func pageToTemplateName(template Template) string {
	switch template {
	case LoginPage:
		return "login_user.html"
	case RegisterPage:
		return "register_user.html"
	case DashboardPage:
		return "view_dashboard.html"
	case PersonalFilePage:
		return "view_personal_files.html"
	case SharePage:
		return "view_shares.html"
	case FileRows:
		return "file_rows"
	case LinkSharePage:
		return "view_links.html"
	case LinkSharePasswordPage:
		return "view_password_upload_link.html"
	case LinkShareDetailPage:
		return "view_upload_link.html"
	case LinkShareCreationPage:
		return "create_upload_link.html"
	default:
		return "not_found.html"
	}
}

func (r *Renderer) Render(w http.ResponseWriter, isAuthenticated bool, template Template, title string, data map[string]any) {
	data["Title"] = title
	data["IsAuthenticated"] = isAuthenticated
	data["Template"] = template
	if r.cfg.DebugMode {
		err := r.parseTemplates()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("could not parse templates: %v", err)
			return
		}
	}
	if err := r.tmpl.ExecuteTemplate(w, pageToTemplateName(template), data); err != nil {
		logger.Error("could not render template: %v", err)
		http.Error(w, "template execution error", http.StatusInternalServerError)
	}
}

// Error has to be used in combination with htmx to display error messages in forms.
func (r *Renderer) Error(w http.ResponseWriter, err string) {
	fmt.Fprintf(w, "<p>%s</p>", err)
}

func (r *Renderer) RedirectHTMX(w http.ResponseWriter, dst string) {
	w.Header().Set("HX-Redirect", dst)
	w.WriteHeader(http.StatusNoContent)
}
