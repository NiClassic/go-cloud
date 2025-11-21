package service

import (
	"context"
	"database/sql"
	"errors"
	"github.com/NiClassic/go-cloud/internal/model"
	"github.com/NiClassic/go-cloud/internal/repository"
	"time"
)

var (
	ErrNotOwner    = errors.New("only owner can share")
	ErrInvalidPerm = errors.New("permission must be read or write")
)

type FileShareService struct {
	fileRepo      repository.FileRepository
	fileShareRepo repository.FileShareRepository
}

func NewFileShareService(fr repository.FileRepository, fsr repository.FileShareRepository) *FileShareService {
	return &FileShareService{fr, fsr}
}

func (s *FileShareService) CreateFileShare(ctx context.Context, ownerID, fileID, recipientID int64, perm string, expires *time.Time) (*model.FileShare, error) {
	if perm != "read" && perm != "write" {
		return nil, ErrInvalidPerm
	}

	file, err := s.fileRepo.GetById(ctx, fileID)
	if err != nil {
		return nil, err
	}
	if file.UserID != ownerID {
		return nil, ErrNotOwner
	}
	share := &model.FileShare{
		FileID:       fileID,
		SharedWithID: recipientID,
		Permission:   perm,
	}
	if expires != nil {
		share.ExpiresAt = sql.NullTime{Valid: true, Time: *expires}
	} else {
		share.ExpiresAt = sql.NullTime{Valid: false}
	}
	err = s.fileShareRepo.Create(ctx, share)
	if err != nil {
		return nil, err
	}
	return share, nil
}

func (s *FileShareService) GetByRecipient(ctx context.Context, userID int64) ([]model.FileShare, error) {
	return s.fileShareRepo.GetByRecipient(ctx, userID)
}

func (s *FileShareService) GetSharedFiles(ctx context.Context, userID int64) ([]repository.SharedFile, error) {
	return s.fileShareRepo.GetSharedFilesForRecipient(ctx, userID)
}

func (s *FileShareService) GetByID(ctx context.Context, requestingUserID, fileShareID int64) (*model.FileShare, error) {
	share, err := s.fileShareRepo.GetByID(ctx, fileShareID)
	if err != nil {
		return nil, err
	}

	file, err := s.fileRepo.GetById(ctx, share.FileID)
	if err != nil {
		return nil, err
	}
	if file.UserID != requestingUserID {
		return nil, ErrNotOwner
	}
	return share, nil
}

func (s *FileShareService) DeleteFileShare(ctx context.Context, requestingUserID, fileShareID int64) error {
	share, err := s.fileShareRepo.GetByID(ctx, fileShareID)
	if err != nil {
		return err
	}

	file, err := s.fileRepo.GetById(ctx, share.FileID)
	if err != nil {
		return err
	}
	if file.UserID != requestingUserID {
		return ErrNotOwner
	}
	return s.fileShareRepo.Delete(ctx, fileShareID)
}
