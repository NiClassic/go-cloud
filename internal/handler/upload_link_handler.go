package handler

import (
	"fmt"
	"github.com/NiClassic/go-cloud/internal/logger"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/NiClassic/go-cloud/internal/service"
)

type UploadLinkHandler struct {
	linkService       *service.UploadLinkService
	linkUnlockService *service.LinkUnlockService
	tmpl              *template.Template
}

func NewUploadLinkHandler(ls *service.UploadLinkService, lu *service.LinkUnlockService, tmpl *template.Template) *UploadLinkHandler {
	return &UploadLinkHandler{
		linkService:       ls,
		linkUnlockService: lu,
		tmpl:              tmpl,
	}
}

func (h *UploadLinkHandler) ShowLinks(w http.ResponseWriter, r *http.Request) {
	logger.Request(r)
	links, err := h.linkService.GetAllLinks(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("could not get all links: %v", err)
		return
	}
	Render(w, h.tmpl, true, LinkSharePage, "Upload Links", map[string]any{
		"Links": links,
		"Now":   time.Now(),
	})
}

func (h *UploadLinkHandler) VisitUploadLink(w http.ResponseWriter, r *http.Request) {
	logger.Request(r)
	suffix := strings.TrimPrefix(r.URL.Path, "/links/")
	parts := strings.SplitN(suffix, "/", 2)
	if len(parts) == 0 {
		http.NotFound(w, r)
		logger.Error("invalid file path: %v", r.URL.Path)
		return
	}
	linkToken := parts[0]

	link, err := h.linkService.GetByToken(r.Context(), linkToken)
	if err != nil {
		logger.Error("could not get upload link: %v", err)
		http.Redirect(w, r, "/links/create", http.StatusSeeOther)
		return
	}
	user := ExtractUserOrRedirect(w, r)

	switch r.Method {
	case http.MethodGet:
		if len(parts) == 2 && parts[1] == "auth" {
			Render(w, h.tmpl, true, LinkSharePasswordPage, "Unlock Link", map[string]any{
				"LinkName":  link.Name,
				"LinkToken": link.LinkToken,
			})
			return
		}
		unlocked, err := h.linkUnlockService.HasUnlocked(r.Context(), user.ID, link.ID)
		if err != nil || !unlocked {
			http.Redirect(w, r, fmt.Sprintf("/links/%s/auth", link.LinkToken), http.StatusSeeOther)
			logger.Error("could not visit upload link: %v", err)
			return
		}
		Render(w, h.tmpl, true, LinkShareDetailPage, "View Link", map[string]any{
			"LinkName": link.Name,
		})

	case http.MethodPost:
		if len(parts) != 2 || parts[1] != "auth" {
			http.Error(w, "invalid request", http.StatusMethodNotAllowed)
			logger.Error("invalid request: %v", r.URL.Path)
			return
		}
		password := r.FormValue("password")
		link, err = h.linkService.ValidatePassword(r.Context(), link.LinkToken, password)
		if err != nil {
			logger.Error("could not validate link password: %v", err)
			http.Redirect(w, r, "/links/create", http.StatusSeeOther)
			return
		}
		err = h.linkUnlockService.UnlockLink(r.Context(), user.ID, link.ID, link.ExpiresAt)
		if err != nil {
			logger.Error("could not unlock link: %v", err)
			http.Redirect(w, r, "/links/create", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/links/"+link.LinkToken, http.StatusSeeOther)
	default:
		logger.InvalidMethod(r)
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
	}
}

func (h *UploadLinkHandler) CreateUploadLink(w http.ResponseWriter, r *http.Request) {
	logger.Request(r)
	switch r.Method {
	case http.MethodGet:
		exp := time.Now().Add(time.Hour).Format("2006-01-02T15:04")
		Render(w, h.tmpl, true, LinkShareCreationPage, "Create Link", map[string]any{
			"DefaultExpiresAt": exp,
		})
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			logger.Error("could not parse form: %v", err)
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		exp, err := time.Parse("2006-01-02T15:04", r.Form.Get("expiry"))
		if err != nil {
			logger.Error("invalid date format: %v", err)
			http.Error(w, "Invalid date format", http.StatusBadRequest)
			return
		}
		link, err := h.linkService.CreateUploadLink(r.Context(),
			r.Form.Get("name"),
			r.Form.Get("password"),
			exp,
		)
		if err != nil {
			logger.Error("could not create upload link: %v", err)
			http.Error(w, "failed to create upload link", http.StatusInternalServerError)
			return
		}
		Render(w, h.tmpl, true, LinkShareCreationPage, "Created Link", map[string]any{
			"LinkName":  link.Name,
			"LinkValue": link.LinkToken,
		})
	default:
		logger.InvalidMethod(r)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
