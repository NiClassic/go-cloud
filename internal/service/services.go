package service

import (
	"database/sql"
	"github.com/NiClassic/go-cloud/internal/repository"
	"github.com/NiClassic/go-cloud/internal/storage"
)

type Services struct {
	Auth        *AuthService
	UploadLink  *UploadLinkService
	LinkSession *UploadLinkSessionService
	PFile       *PersonalFileService
}

// InitServices wires all services and repositories together. It is the main
// dependency injection point.
func InitServices(db *sql.DB, st *storage.Storage) *Services {
	userRepo := repository.NewUserRepository(db)
	sessRepo := repository.NewSessionRepository(db)
	linkRepo := repository.NewUploadLinkRepository(db)
	linkSessRepo := repository.NewUploadLinkSessionRepository(db)
	fileRepo := repository.NewPersonalFileRepository(db)

	authSvc := NewAuthService(userRepo, sessRepo)
	linkSvc := NewUploadLinkService(linkRepo)
	linkSessSvc := NewUploadLinkSessionService(linkSessRepo)
	pFileSvc := NewPersonalFileService(st, fileRepo)

	return &Services{
		Auth:        authSvc,
		UploadLink:  linkSvc,
		LinkSession: linkSessSvc,
		PFile:       pFileSvc,
	}
}
