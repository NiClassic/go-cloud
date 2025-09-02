package handler

import (
	"html/template"
	"net/http"
)

func Render(w http.ResponseWriter, tmpl *template.Template, name, title string, data map[string]any) {
	data["Title"] = title
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "template execution error", http.StatusInternalServerError)
	}
}
