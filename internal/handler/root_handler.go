package handler

import (
	"net/http"

	"github.com/NiClassic/go-cloud/internal/service"
)

type RootHandler struct{ svc *service.AuthService }

func NewRootHandler(svc *service.AuthService) *RootHandler {
	return &RootHandler{svc: svc}
}

func (h *RootHandler) Root(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	user, err := h.svc.GetUserBySessionToken(r.Context(), cookie.Value)
	if err != nil || user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/files", http.StatusSeeOther)
}
