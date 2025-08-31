package service

import (
	"context"
	"encoding/hex"
	"github.com/NiClassic/go-cloud/internal/token"
	"time"

	"github.com/NiClassic/go-cloud/internal/model"
	"github.com/NiClassic/go-cloud/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UploadLinkService struct {
	repo *repository.UploadLinkRepository
}

func NewUploadLinkService(r *repository.UploadLinkRepository) *UploadLinkService {
	return &UploadLinkService{repo: r}
}

func (s *UploadLinkService) CreateUploadLink(
	ctx context.Context,
	name string,
	plain string,
	expiresAt time.Time,
) (*model.UploadLink, error) {
	if name == "" || plain == "" {
		return nil, ErrEmptyLinkFields
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	tok, err := generateUploadToken()
	if err != nil {
		return nil, err
	}
	id, err := s.repo.Insert(ctx, string(hash), tok, name, expiresAt)
	if err != nil {
		return nil, err
	}
	return &model.UploadLink{
		ID:             id,
		HashedPassword: string(hash),
		Name:           name,
		CreatedAt:      time.Now().UTC(),
		ExpiresAt:      expiresAt,
		LinkToken:      tok,
	}, nil
}

func (s *UploadLinkService) GetAllLinks(ctx context.Context) ([]*model.UploadLink, error) {
	return s.repo.GetAll(ctx)
}

func (s *UploadLinkService) ValidatePassword(
	ctx context.Context,
	linkToken, plain string,
) (*model.UploadLink, error) {
	ul, err := s.repo.GetByToken(ctx, linkToken)
	if err != nil {
		return nil, ErrLinkNotFound
	}
	if time.Now().After(ul.ExpiresAt) {
		return nil, ErrLinkExpired
	}
	if err := bcrypt.CompareHashAndPassword([]byte(ul.HashedPassword), []byte(plain)); err != nil {
		return nil, ErrInvalidPassword
	}
	return ul, nil
}

func (s *UploadLinkService) GetByToken(ctx context.Context, linkToken string) (*model.UploadLink, error) {
	return s.repo.GetByToken(ctx, linkToken)
}

func generateUploadToken() (string, error) {
	b, err := token.Bytes(32)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
