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

type FileService struct {
	sto       storage.FileManager
	repo      *repository.FileRepository
	user      *repository.UserRepository
	share     repository.FileShareRepository
	converter *path.Converter
}

func NewFileService(sto storage.FileManager, repo *repository.FileRepository, user *repository.UserRepository, share repository.FileShareRepository, c *path.Converter) *FileService {
	return &FileService{sto, repo, user, share, c}
}

func (s *FileService) DownloadOwn(ctx context.Context, username string, userID, fileID int64) (io.ReadCloser, int64, string, error) {
	file, err := s.repo.GetById(ctx, fileID)
	if err != nil {
		return nil, 0, "", err
	}
	if file.UserID != userID {
		return nil, 0, "", ErrNotOwner
	}
	p := s.converter.GetFullFilePath(username, file.Location)
	f, err := os.Open(p)
	return f, file.Size, s.converter.GetBaseName(p), err
}

func (s *FileService) DownloadShared(ctx context.Context, userID, fileID int64) (io.ReadCloser, int64, string, error) {
	share, err := s.share.GetByFileAndRecipient(ctx, fileID, userID)
	if err != nil {
		return nil, 0, "", err
	}
	file, err := s.repo.GetById(ctx, fileID)
	if err != nil {
		return nil, 0, "", err
	}
	p := s.converter.GetFullFilePath(share.SharedByName, file.Location)

	f, err := os.Open(p)
	return f, file.Size, s.converter.GetBaseName(p), err
}

func (s *FileService) GetUserFiles(ctx context.Context, user *model.User) ([]*model.File, error) {
	return s.repo.GetByUser(ctx, user.ID)
}

func (s *FileService) GetFileById(ctx context.Context, id int64) (*model.File, error) {
	return s.repo.GetById(ctx, id)
}

func (s *FileService) StoreFiles(ctx context.Context, user *model.User, reader *multipart.Reader, folderID int64, folderPath string) error {
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
			fileDBPath = s.converter.JoinDBPath(folderPath, part.FileName())
		}

		// Save to storage
		fileReaderForStorage := bytes.NewReader(fileBytes)
		_, hash, _, err := s.sto.SaveFile(user.Username, folderPath, part.FileName(), fileReaderForStorage)
		if err != nil {
			return fmt.Errorf("failed to save file %q to storage: %w", part.FileName(), err)
		}

		// Store in database with DB path format
		if _, err := s.repo.Insert(ctx,
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

func (s *FileService) DeleteFile(ctx context.Context, user *model.User, fileID int64) error {
	file, err := s.repo.GetById(ctx, fileID)
	if err != nil {
		return err
	}

	if file.UserID != user.ID {
		return fmt.Errorf("unauthorized")
	}

	if err := os.Remove(file.Location); err != nil && !os.IsNotExist(err) {
		return err
	}

	return s.repo.Delete(ctx, fileID)
}
