package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/NiClassic/go-cloud/internal/model"
	"github.com/NiClassic/go-cloud/internal/path"
	"github.com/NiClassic/go-cloud/internal/repository"
	"github.com/NiClassic/go-cloud/internal/storage"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type PersonalFileService struct {
	sto       storage.FileManager
	repo      *repository.PersonalFileRepository
	converter *path.Converter
}

func NewPersonalFileService(sto storage.FileManager, repo *repository.PersonalFileRepository, c *path.Converter) *PersonalFileService {
	return &PersonalFileService{sto, repo, c}
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

		defer part.Close()

		if part.FileName() == "" {
			continue // Skip non-file parts
		}

		// Read file content
		var fileContentBuf bytes.Buffer
		n, copyErr := io.Copy(&fileContentBuf, part)
		if copyErr != nil {
			return fmt.Errorf("failed to read content for file %q: %v", part.FileName(), copyErr)
		}

		fileSize := n
		fileBytes := fileContentBuf.Bytes()
		mimeType := http.DetectContentType(fileBytes)

		// Build the file path in DB format
		var fileDBPath string
		if folderPath == "" {
			fileDBPath = part.FileName() // Root folder
		} else {
			fileDBPath = p.converter.JoinDBPath(folderPath, part.FileName())
		}

		// Save to storage
		fileReaderForStorage := bytes.NewReader(fileBytes)
		_, hash, _, err := p.sto.SaveFile(user.Username, folderPath, part.FileName(), fileReaderForStorage)
		if err != nil {
			return fmt.Errorf("failed to save file %q to storage: %w", part.FileName(), err)
		}

		// Store in database with DB path format
		if _, err := p.repo.Insert(ctx,
			part.FileName(),
			mimeType,
			fileDBPath, // Store relative path in DB
			hash,
			user.ID,
			fileSize,
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
