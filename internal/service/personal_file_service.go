package service

import (
	"context"
	"fmt"
	"github.com/NiClassic/go-cloud/internal/model"
	"github.com/NiClassic/go-cloud/internal/repository"
	"github.com/NiClassic/go-cloud/internal/storage"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

type PersonalFileService struct {
	sto  storage.FileManager
	repo *repository.PersonalFileRepository
}

func NewPersonalFileService(sto storage.FileManager, repo *repository.PersonalFileRepository) *PersonalFileService {
	return &PersonalFileService{sto, repo}
}

func (p *PersonalFileService) GetUserFiles(ctx context.Context, user *model.User) ([]*model.File, error) {
	return p.repo.GetByUser(ctx, user.ID)
}

func extractMetadata(part *multipart.Part) (size int64, mimeType string, err error) {
	mimeType = part.Header.Get("Content-Type")
	if mimeType == "" {
		ext := strings.ToLower(filepath.Ext(part.FileName()))
		mimeType = mime.TypeByExtension(ext)
	}
	if cl := part.Header.Get("Content-Length"); cl != "" {
		_, err = fmt.Sscanf(cl, "%d", &size)
		if err != nil {
			return 0, "", err
		}
	}
	return
}

func (p *PersonalFileService) GetFileById(ctx context.Context, id int64) (*model.File, error) {
	return p.repo.GetById(ctx, id)
}

func (p *PersonalFileService) StoreFiles(ctx context.Context, user *model.User, reader *multipart.Reader, folderID int64, folderPath string) error {
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if part.FileName() == "" {
			err := part.Close()
			if err != nil {
				return err
			}
			continue
		}

		func() {
			defer func(part *multipart.Part) {
				err := part.Close()
				if err != nil {
					return
				}
			}(part)

			_, mimeType, err := extractMetadata(part)
			if err != nil {
				return
			}

			dstPath, hash, size, err := p.sto.SaveFile(user.Username, folderPath, part.FileName(), part)
			if err != nil {
				return
			}

			if _, err := p.repo.Insert(ctx,
				filepath.Base(part.FileName()),
				mimeType,
				dstPath,
				hash,
				user.ID,
				size,
				folderID,
			); err != nil {
				return
			}
		}()
	}

	return nil
}

func (p *PersonalFileService) DeleteFile(ctx context.Context, user *model.User, fileID int64) error {
	file, err := p.repo.GetById(ctx, fileID)
	if err != nil {
		return err
	}

	if file.UserID != user.ID {
		return fmt.Errorf("unauthorized")
	}

	if err := os.Remove(file.Location); err != nil && !os.IsNotExist(err) {
		return err
	}

	return p.repo.Delete(ctx, fileID)
}
