package handler

import (
	"html/template"
	"net/http"
)

type DashboardHandler struct{ tmpl *template.Template }

func NewDashboardHandler(tmpl *template.Template) *DashboardHandler {
	return &DashboardHandler{tmpl: tmpl}
}

func (h *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	user := ExtractUserOrRedirect(w, r)
	Render(w, h.tmpl, true, DashboardPage, "Dashboard", map[string]any{
		"Username": user.Username,
	})
}
