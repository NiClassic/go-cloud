package service

import (
	"database/sql"
	"github.com/NiClassic/go-cloud/internal/repository"
	"github.com/NiClassic/go-cloud/internal/storage"
)

type Services struct {
	Auth       *AuthService
	UploadLink *UploadLinkService
	LinkUnlock *LinkUnlockService
	PFile      *PersonalFileService
	Folder     *FolderService
}

// InitServices wires all services and repositories together. It is the main
// dependency injection point.
func InitServices(db *sql.DB, st *storage.Storage) *Services {
	userRepo := repository.NewUserRepository(db)
	sessRepo := repository.NewSessionRepository(db)
	linkRepo := repository.NewUploadLinkRepository(db)
	linkUnlockRepo := repository.NewLinkUnlockRepository(db)
	fileRepo := repository.NewPersonalFileRepository(db)
	folderRepo := repository.NewFolderRepository(db)

	authSvc := NewAuthService(userRepo, sessRepo)
	linkSvc := NewUploadLinkService(linkRepo)
	linkUnlockSvc := NewLinkUnlockService(linkUnlockRepo)
	folderSvc := NewFolderService(folderRepo, fileRepo)
	pFileSvc := NewPersonalFileService(st, fileRepo)

	return &Services{
		Auth:       authSvc,
		UploadLink: linkSvc,
		LinkUnlock: linkUnlockSvc,
		PFile:      pFileSvc,
		Folder:     folderSvc,
	}
}
