package room

import "context"

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

func (s *Service) List(ctx context.Context) ([]Room, error) {
	return s.repository.List(ctx)
}
