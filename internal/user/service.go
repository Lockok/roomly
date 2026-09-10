package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (User, error) {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.FullName = strings.TrimSpace(input.FullName)

	if err := input.Validate(); err != nil {
		return User{}, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(input.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return User{}, fmt.Errorf("hash password: %w", err)
	}

	input.PasswordHash = string(passwordHash)

	return s.repository.Create(ctx, input)
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]User, error) {
	return s.repository.List(ctx, filter)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) UpdateStatus(ctx context.Context, input UpdateStatusInput) (User, error) {
	return s.repository.UpdateStatus(ctx, input)
}
