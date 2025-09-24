package storage

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path"
)

type FileManager interface {
	// GetBaseDir returns the absolute base directory for the user.
	GetBaseDir(username string) string
	// EnsureDir ensures that a folder exists (creates if missing), returns absolute path.
	EnsureDir(username string, folderPath string) (absPath string, err error)
	// SaveFile saves a file to a folder, returning absolute path, SHA256 hash, and size.
	SaveFile(username string, folderPath, filename string, src io.Reader) (absPath, hash string, size int64, err error)
	// OpenFile opens a file for reading (download).
	OpenFile(username string, folderPath, filename string) (io.ReadCloser, error)
	// DeleteFile deletes a file.
	DeleteFile(username string, folderPath, filename string) error
	// MoveFile moves or renames a file.
	MoveFile(username string, oldFolderPath, oldFilename, newFolderPath, newFilename string) (newAbsPath string, err error)
	// DeleteFolder deletes an empty folder (or recursively).
	DeleteFolder(username string, folderPath string) error
}

type IOStorage struct {
	basePath string
}

func NewIOStorage(basePath string) *IOStorage {
	return &IOStorage{basePath}
}

func (s *IOStorage) GetBaseDir(username string) string {
	return s.getUserDir(username)
}

func (s *IOStorage) EnsureDir(username string, folderPath string) (string, error) {
	absPath := s.absFolderPath(username, folderPath)
	if err := os.MkdirAll(absPath, 0o755); err != nil {
		return "", err
	}
	return absPath, nil
}

func (s *IOStorage) SaveFile(username string, folderPath, filename string, src io.Reader) (string, string, int64, error) {
	if _, err := s.EnsureDir(username, folderPath); err != nil {
		return "", "", 0, err
	}

	dstPath := s.absFilePath(username, folderPath, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", "", 0, err
	}
	defer func(dst *os.File) {
		err := dst.Close()
		if err != nil {
			return
		}
	}(dst)

	hasher := sha256.New()
	tee := io.TeeReader(src, hasher)

	size, err := io.Copy(dst, tee)
	if err != nil {
		return "", "", 0, err
	}

	hash := fmt.Sprintf("%x", hasher.Sum(nil))
	return dstPath, hash, size, nil
}

func (s *IOStorage) OpenFile(username string, folderPath, filename string) (io.ReadCloser, error) {
	return os.Open(s.absFilePath(username, folderPath, filename))
}

func (s *IOStorage) DeleteFile(username string, folderPath, filename string) error {
	return os.Remove(s.absFilePath(username, folderPath, filename))
}

func (s *IOStorage) MoveFile(username string, oldFolder, oldName, newFolder, newName string) (string, error) {
	oldPath := s.absFilePath(username, oldFolder, oldName)
	newFolderAbs := s.absFolderPath(username, newFolder)
	if err := os.MkdirAll(newFolderAbs, 0o755); err != nil {
		return "", err
	}

	newPath := s.absFilePath(username, newFolder, newName)
	if err := os.Rename(oldPath, newPath); err != nil {
		return "", err
	}
	return newPath, nil
}

func (s *IOStorage) DeleteFolder(username string, folderPath string) error {
	absPath := s.absFolderPath(username, folderPath)
	return os.RemoveAll(absPath)
}

func (s *IOStorage) getUserDir(username string) string {
	return path.Join(s.basePath, username)
}

func (s *IOStorage) absFolderPath(username string, folderPath string) string {
	return path.Join(s.getUserDir(username), folderPath)
}

func (s *IOStorage) absFilePath(username string, folderPath string, filename string) string {
	return path.Join(s.getUserDir(username), folderPath, path.Base(filename))
}
