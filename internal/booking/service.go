package booking

import (
	"context"
	"errors"

	"github.com/Lockok/roomly/internal/platform/clock"
	"github.com/Lockok/roomly/internal/room"
)

type Service struct {
	repository Repository
	rooms      RoomReader
	clock      clock.Clock
}

func NewService(
	repository Repository,
	rooms RoomReader,
	clk clock.Clock,
) *Service {
	return &Service{
		repository: repository,
		rooms:      rooms,
		clock:      clk,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Booking, error) {
	if err := input.Validate(s.clock.Now()); err != nil {
		return Booking{}, err
	}

	targetRoom, err := s.rooms.GetByID(ctx, input.RoomID)
	if err != nil {
		if errors.Is(err, room.ErrNotFound) {
			return Booking{}, ErrRoomNotFound
		}

		return Booking{}, err
	}

	if !targetRoom.IsActive {
		return Booking{}, ErrRoomInactive
	}

	if input.AttendeesCount > targetRoom.Capacity {
		return Booking{}, ErrCapacityExceeded
	}

	return s.repository.Create(ctx, input)
}
