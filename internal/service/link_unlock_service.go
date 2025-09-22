package service

import (
	"context"
	"errors"
	"github.com/NiClassic/go-cloud/internal/model"
	"github.com/NiClassic/go-cloud/internal/repository"
	"time"
)

var (
	ErrLinkUnlockInvalid = errors.New("the link unlock has been invalidated")
	ErrLinkUnlockExpired = errors.New("the link unlock has expired")
)

type LinkUnlockService struct {
	repo *repository.LinkUnlockRepository
}

func NewLinkUnlockService(repo *repository.LinkUnlockRepository) *LinkUnlockService {
	return &LinkUnlockService{repo}
}

func (l *LinkUnlockService) GetUserLinks(ctx context.Context, userID int64) ([]*model.LinkUnlock, error) {
	return l.repo.GetByUser(ctx, userID)
}

func (l *LinkUnlockService) UnlockLink(ctx context.Context, userID, uploadLinkID int64, expiry time.Time) error {
	_, err := l.repo.Insert(ctx, userID, uploadLinkID, expiry)
	return err
}

func (l *LinkUnlockService) HasUnlocked(ctx context.Context, userID, uploadLinkID int64) (bool, error) {
	link, err := l.repo.GetByUserAndUploadLinkID(ctx, userID, uploadLinkID)
	if err != nil || link == nil {
		return false, err
	}
	if time.Now().After(link.Expiry) {
		return false, ErrLinkUnlockExpired
	}
	if !link.Valid {
		return false, ErrLinkUnlockInvalid
	}
	return true, nil
}
