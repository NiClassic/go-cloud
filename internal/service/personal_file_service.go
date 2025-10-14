package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/NiClassic/go-cloud/internal/logger"
	"github.com/NiClassic/go-cloud/internal/model"
	"github.com/NiClassic/go-cloud/internal/repository"
	"github.com/NiClassic/go-cloud/internal/storage"
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
			return fmt.Errorf("failed to get next part: %v", err)
		}

		defer func(partToClose *multipart.Part) {
			if closeErr := partToClose.Close(); closeErr != nil {
				logger.Error("warning: failed to close multipart part for %q: %v\n", partToClose.FileName(), closeErr)
			}
		}(part)

		if part.FileName() == "" {
			// Skip non-file parts
			continue
		}

		var fileContentBuf bytes.Buffer
		n, copyErr := io.Copy(&fileContentBuf, part)
		if copyErr != nil {
			return fmt.Errorf("failed to read content for file %q: %v", part.FileName(), copyErr)
		}

		fileSize := n
		fileBytes := fileContentBuf.Bytes()

		mimeType := http.DetectContentType(fileBytes)

		fileReaderForStorage := bytes.NewReader(fileBytes)

		dstPath, hash, _, err := p.sto.SaveFile(user.Username, folderPath, part.FileName(), fileReaderForStorage)
		if err != nil {
			return fmt.Errorf("failed to save file %q to storage: %w", part.FileName(), err)
		}
		if _, err := p.repo.Insert(ctx,
			filepath.Base(part.FileName()),
			mimeType, dstPath,
			hash, // Use hash from storage (or `fileHash` if you prefer service to own it)
			user.ID,
			fileSize, // Use size from storage (or `fileSize` if you prefer service to own it)
			folderID,
		); err != nil {
			return fmt.Errorf("failed to insert file record for %q into database: %w", part.FileName(), err)
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
