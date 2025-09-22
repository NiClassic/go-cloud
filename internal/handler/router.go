package handler

import (
	"github.com/NiClassic/go-cloud/internal/middleware"
	"github.com/NiClassic/go-cloud/internal/service"
	"github.com/NiClassic/go-cloud/internal/storage"
	"html/template"
	"net/http"
)

func New(services *service.Services, st *storage.Storage, tmpl *template.Template) *http.ServeMux {
	authH := NewAuthHandler(services.Auth, tmpl, services.Folder)
	rootH := NewRootHandler(services.Auth)
	uploadH := NewUploadLinkHandler(services.UploadLink, services.LinkUnlock, tmpl)
	pFileH := NewPersonalFileUploadHandler(tmpl, st, services.PFile, services.Folder)
	//folderH := NewFolderHandler(services.Folder, services.PFile)

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	guest := middleware.NewGuestOnly(services.Auth)
	auth := middleware.NewSessionValidator(services.Auth)

	// Authentication routes
	mux.Handle("/register", middleware.Recover(guest.WithoutAuth(http.HandlerFunc(authH.Register))))
	mux.Handle("/login", middleware.Recover(guest.WithoutAuth(http.HandlerFunc(authH.Login))))
	mux.Handle("/logout", middleware.Recover(auth.WithAuth(http.HandlerFunc(authH.Logout))))

	// Upload link routes
	mux.Handle("/links/create", middleware.Recover(auth.WithAuth(http.HandlerFunc(uploadH.CreateUploadLink))))
	mux.Handle("/links", middleware.Recover(auth.WithAuth(http.HandlerFunc(uploadH.ShowLinks))))
	mux.Handle("/links/", middleware.Recover(auth.WithAuth(http.HandlerFunc(uploadH.VisitUploadLink))))

	// File management routes
	mux.Handle("/files", middleware.Recover(auth.WithAuth(http.HandlerFunc(pFileH.RedirectNoTrailingSlash))))
	mux.Handle("/files/", middleware.Recover(auth.WithAuth(http.HandlerFunc(pFileH.ListFiles))))
	mux.Handle("/download/", middleware.Recover(auth.WithAuth(http.HandlerFunc(pFileH.DownloadFile))))
	mux.Handle("/files/upload/", middleware.Recover(auth.WithAuth(http.HandlerFunc(pFileH.UploadFiles))))

	// Root route
	mux.Handle("/", middleware.Recover(http.HandlerFunc(rootH.Root)))

	return mux
}
