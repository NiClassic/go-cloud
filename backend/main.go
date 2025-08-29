package main

import (
	"database/sql"
	"github.com/NiClassic/go-cloud/internal/handler"
	"github.com/NiClassic/go-cloud/internal/middleware"
	"github.com/NiClassic/go-cloud/internal/repository"
	"github.com/NiClassic/go-cloud/internal/service"
	"html/template"
	"log"
	_ "modernc.org/sqlite"
	"net/http"
)

func main() {
	db, err := sql.Open("sqlite", "../data/storage.db")
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewUserRepository(db)
	s := repository.NewSessionRepository(db)
	u := repository.NewLinkTokenRepository(db)
	a := repository.NewUploadLinkSessionRepository(db)
	svc := service.NewAuthService(repo, s)
	uSvc := service.NewUploadLinkService(u)
	uuSvc := service.NewUploadLinkSessionService(a)

	tmpl := template.Must(template.ParseGlob("templates/*.html"))

	authService := service.NewAuthService(repo, s)
	auth := middleware.NewSessionValidator(authService)
	guest := middleware.NewGuestOnly(authService)

	authHandler := handler.NewAuthHandler(svc, tmpl)
	dashboardHandler := handler.NewDashboardHandler(tmpl)
	rootHandler := handler.NewRootHandler(svc)

	uploadLinkHandler := handler.NewUploadLinkHandler(uSvc, uuSvc, tmpl)

	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.Handle("/register", guest.WithoutAuth(http.HandlerFunc(authHandler.Register)))

	mux.Handle("/login", guest.WithoutAuth(http.HandlerFunc(authHandler.Login)))

	mux.Handle("/logout", auth.WithAuth(http.HandlerFunc(authHandler.Logout)))

	mux.Handle("/dashboard", auth.WithAuth(http.HandlerFunc(dashboardHandler.Dashboard)))

	mux.Handle("/links/create", auth.WithAuth(http.HandlerFunc(uploadLinkHandler.CreateUploadLink)))
	mux.Handle("/links", auth.WithAuth(http.HandlerFunc(uploadLinkHandler.ShowLinks)))
	mux.Handle("/links/", auth.WithAuth(http.HandlerFunc(uploadLinkHandler.VisitUploadLink)))

	mux.HandleFunc("/", rootHandler.Root)

	log.Println("listening on :8080")
	http.ListenAndServe(":8080", mux)
}
