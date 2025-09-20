package main

import (
	"database/sql"
	"github.com/NiClassic/go-cloud/config"
	"github.com/NiClassic/go-cloud/internal/handler"
	"github.com/NiClassic/go-cloud/internal/logger"
	"github.com/NiClassic/go-cloud/internal/middleware"
	"github.com/NiClassic/go-cloud/internal/repository"
	"github.com/NiClassic/go-cloud/internal/service"
	"github.com/NiClassic/go-cloud/internal/storage"
	_ "modernc.org/sqlite"
	"net/http"
)

func main() {
	config.Init()

	logger.Init(config.Debug)
	db, err := sql.Open("sqlite", "data/storage.db")
	if err != nil {
		logger.Fatal("could not open database: %v", err)
	}
	if err = db.Ping(); err != nil {
		logger.Fatal("could not ping database: %v", err)
	}
	logger.Info("successfully connected to database")
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Fatal("could not close database: %v", err)
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
		logger.Fatal("could not parse templates: %v", err)
	}

	authH := handler.NewAuthHandler(authSvc, tmpl)
	rootH := handler.NewRootHandler(authSvc)
	uploadH := handler.NewUploadLinkHandler(linkSvc, linkSessSvc, tmpl)
	pFileH := handler.NewPersonalFileUploadHandler(tmpl, storage, pFileSvc)

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	guest := middleware.NewGuestOnly(authSvc)
	auth := middleware.NewSessionValidator(authSvc)

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

	logger.Info("listening on :8080 (Debug Mode=%v)", config.Debug)
	if err = http.ListenAndServe(":8080", mux); err != nil {
		logger.Fatal("could not run server: %v", err)
	}
}
