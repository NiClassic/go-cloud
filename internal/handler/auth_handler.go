package handler

import (
	"github.com/NiClassic/go-cloud/internal/logger"
	"github.com/NiClassic/go-cloud/internal/storage"
	"html/template"
	"net/http"

	"github.com/NiClassic/go-cloud/internal/service"
)

type AuthHandler struct {
	svc           *service.AuthService
	tmpl          *template.Template
	folderService *service.FolderService
	st            storage.FileManager
}

func NewAuthHandler(svc *service.AuthService, tmpl *template.Template, folderService *service.FolderService, st storage.FileManager) *AuthHandler {
	return &AuthHandler{svc: svc, tmpl: tmpl, folderService: folderService, st: st}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	logger.Request(r)
	switch r.Method {
	case http.MethodGet:
		Render(w, h.tmpl, false, LoginPage, "Login", map[string]any{})
	case http.MethodPost:
		username := r.FormValue("username")
		password := r.FormValue("password")

		user, err := h.svc.Authenticate(r.Context(), username, password)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			logger.Error("invalid credentials: %v", err)
			return
		}

		token, err := h.svc.RegisterSession(r.Context(), user)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			logger.Error("internal server error: %v", err)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
		})
		http.Redirect(w, r, "/files", http.StatusSeeOther)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.InvalidMethod(r)
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	logger.Request(r)
	switch r.Method {
	case http.MethodGet:
		Render(w, h.tmpl, false, RegisterPage, "Register", map[string]any{})
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			logger.Error("invalid form: %v", err)
			return
		}
		username := r.FormValue("username")

		userID, err := h.svc.Register(r.Context(), username, r.Form.Get("password"))
		if err != nil {
			http.Error(w, "could not create user", http.StatusInternalServerError)
			logger.Error("could not create user: %v", err)
			return
		}
		if _, err = h.folderService.CreateFolder(r.Context(), userID, username, -1, "/", "/"); err != nil {
			http.Error(w, "could not create folder in db", http.StatusInternalServerError)
			logger.Error("could not create folder in db: %v", err)
			return
		}
		logger.Info("user created: %v, user folder created", r.Form.Get("username"))
		http.Redirect(w, r, "/files", http.StatusSeeOther)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.InvalidMethod(r)
	}
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	logger.Request(r)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.InvalidMethod(r)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err == nil {
		_ = h.svc.DestroySession(r.Context(), cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
