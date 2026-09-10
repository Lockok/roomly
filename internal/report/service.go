package report

import "context"

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) RoomUsage(ctx context.Context, input RoomUsageInput) ([]RoomUsage, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	return s.repository.RoomUsage(ctx, input)
}
