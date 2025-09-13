package main

import (
	"database/sql"
	"github.com/NiClassic/go-cloud/config"
	"github.com/NiClassic/go-cloud/internal/handler"
	"github.com/NiClassic/go-cloud/internal/middleware"
	"github.com/NiClassic/go-cloud/internal/repository"
	"github.com/NiClassic/go-cloud/internal/service"
	"github.com/NiClassic/go-cloud/internal/storage"
	"log"
	_ "modernc.org/sqlite"
	"net/http"
)

func main() {
	db, err := sql.Open("sqlite", "data/storage.db")
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("Error closing sqlite connection: %v", err)
		}
	}(db)

	storage := storage.NewStorage("/home/nico/Code/go/go-cloud/data")

	userRepo := repository.NewUserRepository(db)
	sessRepo := repository.NewSessionRepository(db)
	linkRepo := repository.NewUploadLinkRepository(db)
	linkSessRepo := repository.NewUploadLinkSessionRepository(db)
	fileRepo := repository.NewPersonalFileRepository(db)

	authSvc := service.NewAuthService(userRepo, sessRepo)
	linkSvc := service.NewUploadLinkService(linkRepo)
	linkSessSvc := service.NewUploadLinkSessionService(linkSessRepo)
	pFileSvc := service.NewPersonalFileService(storage, fileRepo)

	tmpl, err := handler.ParseTemplates()
	if err != nil {
		panic(err)
	}

	authH := handler.NewAuthHandler(authSvc, tmpl)
	dashH := handler.NewDashboardHandler(tmpl)
	rootH := handler.NewRootHandler(authSvc)
	uploadH := handler.NewUploadLinkHandler(linkSvc, linkSessSvc, tmpl)
	pFileH := handler.NewPersonalFileUploadHandler(tmpl, storage, pFileSvc)

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	guest := middleware.NewGuestOnly(authSvc)
	auth := middleware.NewSessionValidator(authSvc)

	mux.Handle("/register", guest.WithoutAuth(http.HandlerFunc(authH.Register)))
	mux.Handle("/login", guest.WithoutAuth(http.HandlerFunc(authH.Login)))
	mux.Handle("/logout", auth.WithAuth(http.HandlerFunc(authH.Logout)))
	mux.Handle("/dashboard", auth.WithAuth(http.HandlerFunc(dashH.Dashboard)))
	mux.Handle("/links/create", auth.WithAuth(http.HandlerFunc(uploadH.CreateUploadLink)))
	mux.Handle("/links", auth.WithAuth(http.HandlerFunc(uploadH.ShowLinks)))
	mux.Handle("/links/", auth.WithAuth(http.HandlerFunc(uploadH.VisitUploadLink)))
	mux.Handle("/files", auth.WithAuth(http.HandlerFunc(pFileH.ListFiles)))
	mux.Handle("/files/", auth.WithAuth(http.HandlerFunc(pFileH.DownloadFile)))
	mux.Handle("/files/upload", auth.WithAuth(http.HandlerFunc(pFileH.UploadFiles)))
	mux.HandleFunc("/", rootH.Root)

	config.Init()

	log.Printf("listening on :8080 (Debug Mode=%v)\n", config.Debug)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
