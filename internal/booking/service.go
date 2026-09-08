package booking

import (
	"context"
	"errors"

	"github.com/google/uuid"

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

func (s *Service) List(ctx context.Context, filter ListFilter) ([]Booking, error) {
	if err := filter.Validate(); err != nil {
		return nil, err
	}

	return s.repository.List(ctx, filter)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (Booking, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, input UpdateInput) (Booking, error) {
	now := s.clock.Now()

	if err := input.Validate(now); err != nil {
		return Booking{}, err
	}

	current, err := s.repository.GetByID(ctx, input.ID)
	if err != nil {
		return Booking{}, err
	}

	if current.Status == StatusCancelled {
		return Booking{}, ErrAlreadyCancelled
	}

	if !current.StartsAt.After(now) {
		return Booking{}, ErrBookingStarted
	}

	// Compute the effective time range mixing new and current values so we
	// can validate partial time updates (e.g. only starts_at provided).
	effectiveStart := current.StartsAt
	if input.StartsAt.Set {
		effectiveStart = *input.StartsAt.Value
	}

	effectiveEnd := current.EndsAt
	if input.EndsAt.Set {
		effectiveEnd = *input.EndsAt.Value
	}

	if input.StartsAt.Set || input.EndsAt.Set {
		if err := validateTimeRange(effectiveStart, effectiveEnd, now); err != nil {
			return Booking{}, err
		}
	}

	// If attendees count changes, re-check it against the room capacity.
	if input.AttendeesCount.Set {
		targetRoom, err := s.rooms.GetByID(ctx, current.RoomID)
		if err != nil {
			if errors.Is(err, room.ErrNotFound) {
				return Booking{}, ErrRoomNotFound
			}

			return Booking{}, err
		}

		if *input.AttendeesCount.Value > targetRoom.Capacity {
			return Booking{}, ErrCapacityExceeded
		}
	}

	return s.repository.Update(ctx, input)
}

func (s *Service) Cancel(ctx context.Context, input CancelInput) (Booking, error) {
	current, err := s.repository.GetByID(ctx, input.ID)
	if err != nil {
		return Booking{}, err
	}

	if current.Status == StatusCancelled {
		return Booking{}, ErrAlreadyCancelled
	}

	if !current.StartsAt.After(s.clock.Now()) {
		return Booking{}, ErrBookingStarted
	}

	return s.repository.Cancel(ctx, input)
}
