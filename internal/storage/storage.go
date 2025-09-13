package storage

import (
	"os"
	"path"
)

type Storage struct {
	basePath string
}

func NewStorage(basePath string) *Storage {
	return &Storage{basePath: basePath}
}

func (h *Storage) CreateBaseDirIfAbsent(username string) error {
	fullPath := path.Join(h.basePath, username)
	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return os.MkdirAll(fullPath, 0750)
	}
	return err
}

func (h *Storage) GetBaseDirForUser(username string) string {
	fullPath := path.Join(h.basePath, username)
	return fullPath
}
