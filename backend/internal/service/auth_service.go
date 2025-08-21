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
)

type AuthService struct {
	u *repository.UserRepository
	s *repository.SessionRepository
}

func NewAuthService(u *repository.UserRepository, s *repository.SessionRepository) *AuthService {
	return &AuthService{u: u, s: s}
}

func (a *AuthService) GetUserBySessionToken(ctx context.Context, token string) (*model.User, error) {
	session, err := a.s.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if !session.Valid {
		return nil, errors.New("session invalid")
	}

	return a.u.GetByID(ctx, session.UserID)
}

func generateToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(b)
}

func (a *AuthService) RegisterSession(ctx context.Context, user *model.User) (string, error) {
	token := generateToken()
	_, err := a.s.Insert(ctx, user.ID, token)
	return token, err
}

func (a *AuthService) ValidateSession(ctx context.Context, token string) (bool, error) {
	session, err := a.s.GetByToken(ctx, token)
	return session.Valid, err
}

func (a *AuthService) Register(ctx context.Context, username, plainPassword string) (int64, error) {
	if username == "" || plainPassword == "" {
		return 0, errors.New("username and password required")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	return a.u.Insert(ctx, username, string(hashed))
}

func (a *AuthService) Authenticate(ctx context.Context, username, plainPassword string) (*model.User, error) {
	user, err := a.u.GetByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(plainPassword))
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	return user, nil
}

func (a *AuthService) DestroySession(ctx context.Context, token string) error {
	return a.s.Invalidate(ctx, token)
}
