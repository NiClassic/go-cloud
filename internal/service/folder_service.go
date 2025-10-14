package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/NiClassic/go-cloud/internal/model"
	"github.com/NiClassic/go-cloud/internal/repository"
	"github.com/NiClassic/go-cloud/internal/storage"
)

var (
	ErrFolderNotFound      = errors.New("folder not found")
	ErrInvalidFolderName   = errors.New("invalid folder name")
	ErrFolderAlreadyExists = errors.New("folder already exists")
	ErrCannotMoveToChild   = errors.New("cannot folder to its own child")
	ErrFileNotFound        = errors.New("file not found")
)

type FolderService struct {
	folderRepo *repository.FolderRepository
	fileRepo   *repository.PersonalFileRepository
	st         storage.FileManager
}

func NewFolderService(folderRepo *repository.FolderRepository, fileRepo *repository.PersonalFileRepository, st storage.FileManager) *FolderService {
	return &FolderService{folderRepo, fileRepo, st}
}

func (s *FolderService) CreateFolder(ctx context.Context, userID int64, username string, parentID int64, name, path string) (*model.Folder, error) {
	if name == "" {
		return nil, ErrInvalidFolderName
	}

	if parentID != -1 {
		parent, err := s.folderRepo.GetByID(ctx, parentID)
		if err != nil {
			return nil, ErrFolderNotFound
		}
		if parent.UserID != userID {
			return nil, ErrFolderNotFound
		}
	}
	cleaned := strings.Trim(fmt.Sprintf("/%s%s", username, path), "/")

	var id int64
	var err error
	if parentID == -1 {
		id, err = s.folderRepo.Insert(ctx, userID, nil, name, cleaned)
	} else {
		id, err = s.folderRepo.Insert(ctx, userID, &parentID, name, cleaned)
	}
	if err != nil {
		return nil, err
	}

	_, err = s.st.EnsureDir(username, path)
	if err != nil {
		return nil, err
	}
	return s.folderRepo.GetByID(ctx, id)
}

func (s *FolderService) GetById(ctx context.Context, userID, folderID int64) (*model.Folder, error) {
	folder, err := s.folderRepo.GetByID(ctx, folderID)
	if err != nil {
		return nil, err
	}

	if folder.UserID != userID {
		return nil, ErrFolderNotFound
	}
	return folder, nil
}

func (s *FolderService) GetByPath(ctx context.Context, userID int64, username string, path string) (*model.Folder, error) {
	cleaned := strings.Trim(fmt.Sprintf("/%s%s", username, path), "/")
	folder, err := s.folderRepo.GetByPath(ctx, cleaned)
	if err != nil {
		return nil, err
	}

	if folder.UserID != userID {
		return nil, ErrFolderNotFound
	}
	return folder, nil
}

func (s *FolderService) GetFolderContents(ctx context.Context, userID int64, folderID int64) ([]*model.Folder, []*model.File, error) {
	folder, err := s.folderRepo.GetByID(ctx, folderID)
	if err != nil {
		return nil, nil, ErrFolderNotFound
	}
	if folder.UserID != userID {
		return nil, nil, ErrFolderNotFound
	}

	folders, err := s.folderRepo.GetByUserAndParent(ctx, userID, folderID)
	if err != nil {
		return nil, nil, err
	}

	files, err := s.fileRepo.GetByUserAndFolder(ctx, userID, folderID)
	if err != nil {
		return nil, nil, err
	}

	return folders, files, nil
}

func (s *FolderService) MoveFolder(ctx context.Context, userID, folderID int64, newParentID int64) error {
	folder, err := s.folderRepo.GetByID(ctx, folderID)
	if err != nil {
		return err
	}
	if folder.UserID != userID {
		return ErrFolderNotFound
	}

	parent, err := s.folderRepo.GetByID(ctx, newParentID)
	if err != nil {
		return ErrFolderNotFound
	}
	if parent.UserID != userID {
		return ErrFolderNotFound
	}

	return s.folderRepo.UpdateParent(ctx, folderID, newParentID)
}

func (s *FolderService) MoveFile(ctx context.Context, userID, fileID int64, folderID int64) error {
	file, err := s.fileRepo.GetById(ctx, fileID)
	if err != nil {
		return err
	}
	if file.UserID != userID {
		return ErrFileNotFound
	}

	folder, err := s.folderRepo.GetByID(ctx, folderID)
	if err != nil {
		return ErrFolderNotFound
	}
	if folder.UserID != userID {
		return ErrFolderNotFound
	}

	return s.fileRepo.UpdateFolder(ctx, fileID, folderID)
}

func (s *FolderService) DeleteFolder(ctx context.Context, userID, folderID int64) error {
	folder, err := s.folderRepo.GetByID(ctx, folderID)
	if err != nil {
		return err
	}
	if folder.UserID != userID {
		return ErrFolderNotFound
	}

	return s.folderRepo.Delete(ctx, folderID)
}
