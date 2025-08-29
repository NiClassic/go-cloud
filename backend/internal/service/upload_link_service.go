package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/NiClassic/go-cloud/internal/model"
	"github.com/NiClassic/go-cloud/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

type UploadLinkService struct {
	r *repository.UploadLinkRepository
}

func NewUploadLinkService(r *repository.UploadLinkRepository) *UploadLinkService {
	return &UploadLinkService{r: r}
}

func generateUploadToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(b)
}

func (s *UploadLinkService) CreateUploadLink(
	ctx context.Context,
	name string,
	plainPassword string,
	expiresAt time.Time,
) (*model.UploadLink, error) {
	if name == "" || plainPassword == "" {
		return nil, errors.New("name and password required")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	token := generateUploadToken()

	id, err := s.r.Insert(ctx, string(hashed), token, name, expiresAt)
	if err != nil {
		return nil, err
	}

	return &model.UploadLink{
		ID:             id,
		HashedPassword: string(hashed),
		Name:           name,
		CreatedAt:      time.Now(),
		ExpiresAt:      expiresAt,
		LinkToken:      token,
	}, nil
}

func (s *UploadLinkService) GetAllLinks(ctx context.Context) ([]*model.UploadLink, error) {
	return s.r.GetAll(ctx)
}

func (s *UploadLinkService) ValidatePassword(
	ctx context.Context,
	linkToken string,
	plainPassword string,
) (*model.UploadLink, error) {
	uploadLink, err := s.r.GetByToken(ctx, linkToken)
	if err != nil {
		return nil, errors.New("upload link not found")
	}

	if time.Now().After(uploadLink.ExpiresAt) {
		return nil, errors.New("upload link expired")
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(uploadLink.HashedPassword),
		[]byte(plainPassword),
	)
	if err != nil {
		return nil, errors.New("invalid password")
	}

	return uploadLink, nil
}

func (s *UploadLinkService) GetByToken(
	ctx context.Context,
	linkToken string,
) (*model.UploadLink, error) {
	uploadLink, err := s.r.GetByToken(ctx, linkToken)
	if err != nil {
		return nil, err
	}
	//TODO: Add check if link is expired
	//if time.Now().After(uploadLink.ExpiresAt) {
	//	return nil, errors.New("upload link expired")
	//}
	return uploadLink, nil
}
