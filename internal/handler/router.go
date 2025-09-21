package handler

import (
	"github.com/NiClassic/go-cloud/internal/middleware"
	"github.com/NiClassic/go-cloud/internal/service"
	"github.com/NiClassic/go-cloud/internal/storage"
	"html/template"
	"net/http"
)

func New(services *service.Services, st *storage.Storage, tmpl *template.Template) *http.ServeMux {
	authH := NewAuthHandler(services.Auth, tmpl)
	rootH := NewRootHandler(services.Auth)
	uploadH := NewUploadLinkHandler(services.UploadLink, services.LinkSession, tmpl)
	pFileH := NewPersonalFileUploadHandler(tmpl, st, services.PFile)

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	guest := middleware.NewGuestOnly(services.Auth)
	auth := middleware.NewSessionValidator(services.Auth)

	mux.Handle("/register", middleware.Recover(guest.WithoutAuth(http.HandlerFunc(authH.Register))))
	mux.Handle("/login", middleware.Recover(guest.WithoutAuth(http.HandlerFunc(authH.Login))))
	mux.Handle("/logout", middleware.Recover(auth.WithAuth(http.HandlerFunc(authH.Logout))))
	mux.Handle("/links/create", middleware.Recover(auth.WithAuth(http.HandlerFunc(uploadH.CreateUploadLink))))
	mux.Handle("/links", middleware.Recover(auth.WithAuth(http.HandlerFunc(uploadH.ShowLinks))))
	mux.Handle("/links/", middleware.Recover(auth.WithAuth(http.HandlerFunc(uploadH.VisitUploadLink))))
	mux.Handle("/files", middleware.Recover(auth.WithAuth(http.HandlerFunc(pFileH.ListFiles))))
	mux.Handle("/files/", middleware.Recover(auth.WithAuth(http.HandlerFunc(pFileH.DownloadFile))))
	mux.Handle("/files/upload", middleware.Recover(auth.WithAuth(http.HandlerFunc(pFileH.UploadFiles))))
	mux.Handle("/", middleware.Recover(http.HandlerFunc(rootH.Root)))

	return mux
}
