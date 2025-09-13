package service

import (
	"github.com/NiClassic/go-cloud/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
	s    *repository.SessionRepository
}

func NewUserService(repo *repository.UserRepository, s *repository.SessionRepository) *UserService {
	return &UserService{repo: repo, s: s}
}
