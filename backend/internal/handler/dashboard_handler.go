package handler

import (
	"github.com/NiClassic/go-cloud/internal/middleware"
	"github.com/NiClassic/go-cloud/internal/model"
	"html/template"
	"log"
	"net/http"
)

type DashboardHandler struct {
	tmpl *template.Template
}

func NewDashboardHandler(tmpl *template.Template) *DashboardHandler {
	return &DashboardHandler{tmpl: tmpl}
}

func (h *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.ContextUserKey).(*model.User)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	err := h.tmpl.ExecuteTemplate(w, "dashboard.html", map[string]any{
		"Title":    "Dashboard | Go-Cloud",
		"Username": user.Username,
	})
	if err != nil {
		log.Fatal(err)
	}
}
