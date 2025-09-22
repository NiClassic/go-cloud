package service

import (
	"context"
	"crypto/sha256"
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
	sto  *storage.Storage
	repo *repository.PersonalFileRepository
}

func NewPersonalFileService(sto *storage.Storage, repo *repository.PersonalFileRepository) *PersonalFileService {
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

func (p *PersonalFileService) StoreFiles(ctx context.Context, user *model.User, reader *multipart.Reader, folderID int64) error {
	err := p.sto.CreateBaseDirIfAbsent(user.Username)
	if err != nil {
		return err
	}

	baseUserPath := p.sto.GetBaseDirForUser(user.Username)

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if part.FileName() == "" {
			continue
		}

		_, mimeType, err := extractMetadata(part)
		if err != nil {
			return err
		}

		safeName := filepath.Base(part.FileName())
		dstPath := filepath.Join(baseUserPath, safeName)

		dst, err := os.Create(dstPath)
		if err != nil {
			return err
		}

		hasher := sha256.New()
		tee := io.TeeReader(part, hasher)

		written, err := io.Copy(dst, tee)
		if err != nil {
			err := dst.Close()
			if err != nil {
				return err
			}
			return err
		}

		hash := fmt.Sprintf("%x", hasher.Sum(nil))

		if _, err = p.repo.Insert(ctx, safeName, mimeType, dstPath, hash, user.ID, written, folderID); err != nil {
			err := dst.Close()
			if err != nil {
				return err
			}
			return err
		}

		if err = part.Close(); err != nil {
			err := dst.Close()
			if err != nil {
				return err
			}
			return err
		}

		if err = dst.Close(); err != nil {
			return err
		}
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
