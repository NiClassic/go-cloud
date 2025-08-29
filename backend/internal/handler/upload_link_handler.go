package handler

import (
	"context"
	"fmt"
	"github.com/NiClassic/go-cloud/internal/middleware"
	"github.com/NiClassic/go-cloud/internal/model"
	"github.com/NiClassic/go-cloud/internal/service"
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
	"time"
)

type UploadLinkHandler struct {
	linkService    *service.UploadLinkService
	sessionService *service.UploadLinkSessionService
	tmpl           *template.Template
}

func NewUploadLinkHandler(linkService *service.UploadLinkService, sessionService *service.UploadLinkSessionService, tmpl *template.Template) *UploadLinkHandler {
	return &UploadLinkHandler{
		linkService:    linkService,
		sessionService: sessionService,
		tmpl:           tmpl,
	}
}

func (u *UploadLinkHandler) ShowLinks(w http.ResponseWriter, r *http.Request) {
	links, err := u.linkService.GetAllLinks(context.TODO())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = u.tmpl.ExecuteTemplate(w, "view_links.html", map[string]any{
		"Title": "Upload Links | Go-Cloud",
		"Links": links,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func (u *UploadLinkHandler) VisitUploadLink(w http.ResponseWriter, r *http.Request) {
	// /links/abc123/auth
	// /links/abc123

	//GET:	/links/abc123 -> Upload form
	//		/links/abc123/auth -> form
	//POST:	/links/abc123/auth -> Enter password
	suffix := strings.TrimPrefix(r.URL.Path, "/links/")
	parts := strings.SplitN(suffix, "/", 2)

	if len(parts) == 0 {
		http.NotFound(w, r)
		return
	}

	linkToken := parts[0]
	link, err := u.linkService.GetByToken(context.TODO(), linkToken)
	if err != nil {
		http.Redirect(w, r, "/links/create", http.StatusSeeOther)
		return
	}
	switch r.Method {
	case http.MethodGet:
		if len(parts) == 2 && parts[1] == "auth" {
			err = u.tmpl.ExecuteTemplate(w, "password_upload_link.html", map[string]any{
				"Title":     "Unlock Link | Go-Cloud",
				"LinkName":  link.Name,
				"LinkToken": link.LinkToken,
			})
			if err != nil {
				log.Fatal(err)
			}
			return
		}
		cookie, err := r.Cookie(link.LinkToken)
		if err != nil {
			http.Redirect(w, r, fmt.Sprintf("/links/%s/auth", link.LinkToken), http.StatusSeeOther)
			return
		}

		if ok, err := u.sessionService.ValidateSession(context.TODO(), cookie.Value); err != nil || !ok {
			http.Redirect(w, r, fmt.Sprintf("/links/%s/auth", link.LinkToken), http.StatusSeeOther)
			return
		}

		err = u.tmpl.ExecuteTemplate(w, "view_upload_link.html", map[string]any{
			"Title":    "View Link | Go-Cloud",
			"LinkName": link.Name,
		})
		if err != nil {
			log.Fatal(err)
		}
	case http.MethodPost:
		if len(parts) != 2 || parts[1] != "auth" {
			http.Error(w, "invalid request", http.StatusMethodNotAllowed)
			return
		}
		password := r.FormValue("password")
		link, err = u.linkService.ValidatePassword(context.TODO(), link.LinkToken, password)
		if err != nil {
			http.Redirect(w, r, "/links/create", http.StatusSeeOther)
			return
		}

		session, err := u.sessionService.RegisterSession(context.TODO(), link, r.Context().Value(middleware.ContextUserKey).(*model.User))
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
		return
	default:
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}
}

func (u *UploadLinkHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		linkToken := path.Base(r.URL.Path)
		if linkToken == "" {
			http.Redirect(w, r, "/upload_link", http.StatusSeeOther)
			return
		}

		link, err := u.linkService.GetByToken(context.TODO(), linkToken)
		if err != nil {
			http.Redirect(w, r, "/upload_link", http.StatusSeeOther)
			return
		}
		err = u.tmpl.ExecuteTemplate(w, "password_upload_link.html", map[string]any{
			"Title":    "Unlock Link | Go-Cloud",
			"LinkName": link.Name,
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (u *UploadLinkHandler) CreateUploadLink(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		expiresAt := time.Now().Add(1 * time.Hour)
		defaultValue := expiresAt.Format("2006-01-02T15:04")
		err := u.tmpl.ExecuteTemplate(w, "create_upload_link.html", map[string]any{
			"Title":            "Create Link | Go-Cloud",
			"DefaultExpiresAt": defaultValue,
		})
		if err != nil {
			log.Fatal(err)
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}

		name := r.Form.Get("name")
		password := r.Form.Get("password")
		expiresAtStr := r.Form.Get("expiry")

		expiresAt, err := time.Parse("2006-01-02T15:04", expiresAtStr)
		if err != nil {
			http.Error(w, "Invalid date format", http.StatusBadRequest)
			return
		}

		link, err := u.linkService.CreateUploadLink(context.TODO(), name, password, expiresAt)
		if err != nil {
			http.Error(w, "failed to create upload link", http.StatusInternalServerError)
			return
		}
		err = u.tmpl.ExecuteTemplate(w, "created_upload_link.html", map[string]any{
			"Title":     "Created Link | Go-Cloud",
			"LinkName":  link.Name,
			"LinkValue": link.LinkToken,
		})
		if err != nil {
			log.Fatal(err)
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
