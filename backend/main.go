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
	svc := service.NewAuthService(repo, s)
	tmpl := template.Must(template.ParseGlob("templates/*.html"))

	authService := service.NewAuthService(repo, s)
	auth := middleware.NewSessionValidator(authService)
	guest := middleware.NewGuestOnly(authService)

	authHandler := handler.NewAuthHandler(svc, tmpl)
	dashboardHandler := handler.NewDashboardHandler(tmpl)
	rootHandler := handler.NewRootHandler(svc)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.Handle("/register", guest.WithoutAuth(http.HandlerFunc(authHandler.Register)))

	http.Handle("/login", guest.WithoutAuth(http.HandlerFunc(authHandler.Login)))

	http.Handle("/logout", auth.WithAuth(http.HandlerFunc(authHandler.Logout)))

	http.Handle("/dashboard", auth.WithAuth(http.HandlerFunc(dashboardHandler.Dashboard)))

	http.HandleFunc("/", rootHandler.Root)

	log.Println("listening on :8080")
	http.ListenAndServe(":8080", nil)
}
