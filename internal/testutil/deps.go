package testutil

import (
	"context"
	"database/sql"
	"github.com/NiClassic/go-cloud/internal/repository"
	"testing"
)

type TestDeps struct {
	DB     *sql.DB
	User   *repository.UserRepository
	Folder *repository.FolderRepository
	File   *repository.FileRepository
	Share  *repository.FileShareRepositoryImpl
	Ctx    context.Context
}

func NewTestDeps(t *testing.T) *TestDeps {
	t.Helper()
	db := SetupTestDB(t)
	return &TestDeps{
		DB:     db,
		User:   repository.NewUserRepository(db),
		Folder: repository.NewFolderRepository(db),
		File:   repository.NewFileRepository(db),
		Share:  repository.NewFileShareRepositoryImpl(db),
		Ctx:    TestContext(t),
	}
}
