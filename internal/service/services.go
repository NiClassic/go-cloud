package service

import (
	"database/sql"
	"github.com/NiClassic/go-cloud/internal/path"
	"github.com/NiClassic/go-cloud/internal/repository"
	"github.com/NiClassic/go-cloud/internal/storage"
)

type Services struct {
	Auth       *AuthService
	UploadLink *UploadLinkService
	LinkUnlock *LinkUnlockService
	PFile      *FileService
	Folder     *FolderService
	FileShare  *FileShareService
}

// InitServices wires all services and repositories together. It is the main
// dependency injection point.
func InitServices(db *sql.DB, st storage.FileManager, c *path.Converter) *Services {
	userRepo := repository.NewUserRepository(db)
	sessRepo := repository.NewSessionRepository(db)
	linkRepo := repository.NewUploadLinkRepository(db)
	linkUnlockRepo := repository.NewLinkUnlockRepository(db)
	fileRepo := repository.NewFileRepository(db)
	folderRepo := repository.NewFolderRepository(db)
	fileShareRepo := repository.NewFileShareRepositoryImpl(db)

	authSvc := NewAuthService(userRepo, sessRepo)
	linkSvc := NewUploadLinkService(linkRepo)
	linkUnlockSvc := NewLinkUnlockService(linkUnlockRepo)
	folderSvc := NewFolderService(folderRepo, fileRepo, st, c)
	pFileSvc := NewFileService(st, fileRepo, userRepo, fileShareRepo, c)
	fileShareSvc := NewFileShareService(*fileRepo, fileShareRepo)

	return &Services{
		Auth:       authSvc,
		UploadLink: linkSvc,
		LinkUnlock: linkUnlockSvc,
		PFile:      pFileSvc,
		Folder:     folderSvc,
		FileShare:  fileShareSvc,
	}
}
