package service

import (
	"context"
	"errors"

	"github.com/DNA-Z/Ignis/internal/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
)

type AuthService interface {
	Register(ctx context.Context, req *models.RegisterRequest) (*models.UserResponse, string, error)
}
