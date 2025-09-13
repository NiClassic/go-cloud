package handler

import (
	"html/template"
	"net/http"

	"github.com/NiClassic/go-cloud/internal/model"
)

type DashboardHandler struct{ tmpl *template.Template }

func NewDashboardHandler(tmpl *template.Template) *DashboardHandler {
	return &DashboardHandler{tmpl: tmpl}
}

func (h *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(UserKey).(*model.User)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	Render(w, h.tmpl, true, "dashboard.html", "Dashboard", map[string]any{
		"Username": user.Username,
	})
}
