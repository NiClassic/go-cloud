package handler

import (
	"context"
	"github.com/NiClassic/go-cloud/internal/service"
	"html/template"
	"log"
	"net/http"
)

type AuthHandler struct {
	svc  *service.AuthService
	tmpl *template.Template
}

func NewAuthHandler(svc *service.AuthService, tmpl *template.Template) *AuthHandler {
	return &AuthHandler{svc: svc, tmpl: tmpl}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		err := h.tmpl.ExecuteTemplate(w, "login.html", map[string]any{
			"Title": "Login | Go-Cloud",
		})
		if err != nil {
			log.Fatal(err)
		}
	case http.MethodPost:
		username := r.FormValue("username")
		password := r.FormValue("password")

		user, err := h.svc.Authenticate(context.TODO(), username, password)
		if err != nil {
			http.Error(w, "Invalid credential", http.StatusUnauthorized)
			return
		}

		token, err := h.svc.RegisterSession(context.TODO(), user)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		//TODO: set expiry
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
		})

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		err := h.tmpl.ExecuteTemplate(w, "register.html", map[string]any{
			"Title": "Register | Go-Cloud",
		})
		if err != nil {
			log.Fatal(err)
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}

		username := r.Form.Get("username")
		password := r.Form.Get("password")

		_, err := h.svc.Register(r.Context(), username, password)
		if err != nil {
			http.Error(w, "could not create user", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusUnauthorized)
			return
		}
		err = h.svc.DestroySession(context.TODO(), cookie.Value)
		if err != nil {
			log.Println(err)
		}
		// Clear the session cookie
		cookie = &http.Cookie{
			Name:     "session_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1, // expire immediately
			HttpOnly: true,
			Secure:   true,
		}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}
