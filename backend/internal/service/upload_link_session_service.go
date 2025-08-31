package service

import (
	"context"
	"github.com/NiClassic/go-cloud/internal/model"
	"github.com/NiClassic/go-cloud/internal/repository"
)

type UploadLinkSessionService struct {
	repo *repository.UploadLinkSessionRepository
}

func NewUploadLinkSessionService(r *repository.UploadLinkSessionRepository) *UploadLinkSessionService {
	return &UploadLinkSessionService{repo: r}
}

func (u *UploadLinkSessionService) RegisterSession(
	ctx context.Context,
	link *model.UploadLink,
	user *model.User,
) (string, error) {
	tok, err := generateToken()
	if err != nil {
		return "", err
	}
	_, err = u.repo.Insert(ctx, user.ID, link.ID, tok)
	return tok, err
}

func (u *UploadLinkSessionService) ValidateSession(ctx context.Context, token string) (bool, error) {
	sess, err := u.repo.GetByToken(ctx, token)
	if err != nil {
		return false, err
	}
	return sess.Valid, nil
}

func (u *UploadLinkSessionService) DestroySession(ctx context.Context, token string) error {
	return u.repo.Invalidate(ctx, token)
}
