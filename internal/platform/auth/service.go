package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Lockok/roomly/internal/user"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user is inactive")
)

type TokenGenerator interface {
	Generate(userID uuid.UUID, role string) (string, error)
}

type UserReader interface {
	GetByEmail(ctx context.Context, email string) (user.User, error)
}

type Service struct {
	users  UserReader
	tokens TokenGenerator
}

func NewService(users UserReader, tokens TokenGenerator) *Service {
	return &Service{
		users:  users,
		tokens: tokens,
	}
}

func (s *Service) Login(ctx context.Context, email, password string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	found, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return "", ErrInvalidCredentials
		}

		return "", fmt.Errorf("get user by email: %w", err)
	}

	if !found.IsActive {
		return "", ErrUserInactive
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(found.PasswordHash),
		[]byte(password),
	); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := s.tokens.Generate(found.ID, string(found.Role))
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return token, nil
}
