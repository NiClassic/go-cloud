package service

import (
	"context"
	"encoding/hex"
	"github.com/NiClassic/go-cloud/internal/token"

	"github.com/NiClassic/go-cloud/internal/model"
	"github.com/NiClassic/go-cloud/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users    *repository.UserRepository
	sessions *repository.SessionRepository
}

func NewAuthService(u *repository.UserRepository, s *repository.SessionRepository) *AuthService {
	return &AuthService{users: u, sessions: s}
}

func (a *AuthService) GetUserBySessionToken(ctx context.Context, token string) (*model.User, error) {
	sess, err := a.sessions.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if !sess.Valid {
		return nil, ErrSessionInvalid
	}
	return a.users.GetByID(ctx, sess.UserID)
}

func (a *AuthService) RegisterSession(ctx context.Context, u *model.User) (string, error) {
	tok, err := generateToken()
	if err != nil {
		return "", err
	}
	_, err = a.sessions.Insert(ctx, u.ID, tok)
	return tok, err
}

func (a *AuthService) ValidateSession(ctx context.Context, token string) (bool, error) {
	sess, err := a.sessions.GetByToken(ctx, token)
	if err != nil {
		return false, err
	}
	return sess.Valid, nil
}

func (a *AuthService) Register(ctx context.Context, username, plain string) (int64, error) {
	if username == "" || plain == "" {
		return 0, ErrEmptyCredentials
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	return a.users.Insert(ctx, username, string(hash))
}

func (a *AuthService) Authenticate(ctx context.Context, username, plain string) (*model.User, error) {
	u, err := a.users.GetByUsername(ctx, username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(plain)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return u, nil
}

func (a *AuthService) DestroySession(ctx context.Context, token string) error {
	return a.sessions.Invalidate(ctx, token)
}

func generateToken() (string, error) {
	b, err := token.Bytes(32)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
