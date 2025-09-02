package handler

import (
	"fmt"
	"github.com/NiClassic/go-cloud/internal/model"
	"github.com/NiClassic/go-cloud/internal/service"
	"html/template"
	"net/http"
	"strings"
	"time"
)

type UploadLinkHandler struct {
	linkService    *service.UploadLinkService
	sessionService *service.UploadLinkSessionService
	tmpl           *template.Template
}

func NewUploadLinkHandler(ls *service.UploadLinkService, ss *service.UploadLinkSessionService, tmpl *template.Template) *UploadLinkHandler {
	return &UploadLinkHandler{
		linkService:    ls,
		sessionService: ss,
		tmpl:           tmpl,
	}
}

func (h *UploadLinkHandler) ShowLinks(w http.ResponseWriter, r *http.Request) {
	links, err := h.linkService.GetAllLinks(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	Render(w, h.tmpl, "view_links.html", "Upload Links | Go-Cloud", map[string]any{
		"Links": links,
	})
}

func (h *UploadLinkHandler) VisitUploadLink(w http.ResponseWriter, r *http.Request) {
	suffix := strings.TrimPrefix(r.URL.Path, "/links/")
	parts := strings.SplitN(suffix, "/", 2)
	if len(parts) == 0 {
		http.NotFound(w, r)
		return
	}
	linkToken := parts[0]

	link, err := h.linkService.GetByToken(r.Context(), linkToken)
	if err != nil {
		http.Redirect(w, r, "/links/create", http.StatusSeeOther)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if len(parts) == 2 && parts[1] == "auth" {
			Render(w, h.tmpl, "password_upload_link.html", "Unlock Link | Go-Cloud", map[string]any{
				"LinkName":  link.Name,
				"LinkToken": link.LinkToken,
			})
			return
		}
		cookie, err := r.Cookie(link.LinkToken)
		if err == nil {
			if ok, _ := h.sessionService.ValidateSession(r.Context(), cookie.Value); ok {
				Render(w, h.tmpl, "view_upload_link.html", "View Link | Go-Cloud", map[string]any{
					"LinkName": link.Name,
				})
				return
			}
		}
		http.Redirect(w, r, fmt.Sprintf("/links/%s/auth", link.LinkToken), http.StatusSeeOther)

	case http.MethodPost:
		if len(parts) != 2 || parts[1] != "auth" {
			http.Error(w, "invalid request", http.StatusMethodNotAllowed)
			return
		}
		password := r.FormValue("password")
		link, err = h.linkService.ValidatePassword(r.Context(), link.LinkToken, password)
		if err != nil {
			http.Redirect(w, r, "/links/create", http.StatusSeeOther)
			return
		}
		user, _ := r.Context().Value(UserKey).(*model.User)
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		session, err := h.sessionService.RegisterSession(r.Context(), link, user)
		if err != nil {
			http.Redirect(w, r, "/links/create", http.StatusSeeOther)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     link.LinkToken,
			Value:    session,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
		})
		http.Redirect(w, r, "/links/"+link.LinkToken, http.StatusSeeOther)
	default:
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
	}
}

func (h *UploadLinkHandler) CreateUploadLink(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		exp := time.Now().Add(time.Hour).Format("2006-01-02T15:04")
		Render(w, h.tmpl, "create_upload_link.html", "Create Link | Go-Cloud", map[string]any{
			"DefaultExpiresAt": exp,
		})
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		exp, err := time.Parse("2006-01-02T15:04", r.Form.Get("expiry"))
		if err != nil {
			http.Error(w, "Invalid date format", http.StatusBadRequest)
			return
		}
		link, err := h.linkService.CreateUploadLink(r.Context(),
			r.Form.Get("name"),
			r.Form.Get("password"),
			exp,
		)
		if err != nil {
			http.Error(w, "failed to create upload link", http.StatusInternalServerError)
			return
		}
		Render(w, h.tmpl, "created_upload_link.html", "Created Link | Go-Cloud", map[string]any{
			"LinkName":  link.Name,
			"LinkValue": link.LinkToken,
		})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
