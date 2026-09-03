package room

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Room, error) {
	if err := input.Validate(); err != nil {
		return Room{}, err
	}

	return s.repository.Create(ctx, input)
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]Room, error) {
	return s.repository.List(ctx, filter)
}

func (s *Service) ListAvailable(ctx context.Context, input AvailabilityInput) ([]Room, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	return s.repository.ListAvailable(ctx, input)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (Room, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, input UpdateInput) (Room, error) {
	if err := input.Validate(); err != nil {
		return Room{}, err
	}

	return s.repository.Update(ctx, input)
}
