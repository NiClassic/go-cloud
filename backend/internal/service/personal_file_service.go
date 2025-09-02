package service

import (
	"github.com/NiClassic/go-cloud/internal/storage"
	"os"
)

type PersonalFileService struct {
	sto *storage.Storage
}

func NewPersonalFileService(sto *storage.Storage) *PersonalFileService {
	return &PersonalFileService{sto}
}

func (p *PersonalFileService) GetUserFiles(username string) ([]string, error) {
	baseUserPath := p.sto.GetBaseDirForUser(username)
	entries, err := os.ReadDir(baseUserPath)
	var files []string
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		files = append(files, e.Name())
	}
	return files, nil
}
