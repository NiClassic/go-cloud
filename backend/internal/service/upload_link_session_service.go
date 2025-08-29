package service

import (
	"context"
	"github.com/NiClassic/go-cloud/internal/model"
	"github.com/NiClassic/go-cloud/internal/repository"
)

type UploadLinkSessionService struct {
	r *repository.UploadLinkSessionRepository
}

func NewUploadLinkSessionService(r *repository.UploadLinkSessionRepository) *UploadLinkSessionService {
	return &UploadLinkSessionService{r: r}
}

func (u *UploadLinkSessionService) RegisterSession(ctx context.Context, uploadLink *model.UploadLink, user *model.User) (string, error) {
	token := generateToken()
	_, err := u.r.Insert(ctx, user.ID, uploadLink.ID, token)
	return token, err
}

func (u *UploadLinkSessionService) ValidateSession(ctx context.Context, token string) (bool, error) {
	session, err := u.r.GetByToken(ctx, token)
	if err != nil {
		return false, err
	}
	return session.Valid, err
}

func (u *UploadLinkSessionService) DestroySession(ctx context.Context, token string) error {
	return u.r.Invalidate(ctx, token)
}
