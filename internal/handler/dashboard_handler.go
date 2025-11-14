package handler

import (
	"github.com/NiClassic/go-cloud/config"
	"net/http"
)

type DashboardHandler struct{ *baseHandler }

func NewDashboardHandler(cfg *config.Config, r *Renderer) *DashboardHandler {
	return &DashboardHandler{baseHandler: newBaseHandler(cfg, r)}
}

func (h *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	user := ExtractUserOrRedirect(w, r)
	h.r.Render(w, true, DashboardPage, "Dashboard", map[string]any{
		"Username": user.Username,
	})
}
