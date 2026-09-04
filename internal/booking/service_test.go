package booking

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Lockok/roomly/internal/room"
	"github.com/google/uuid"
)

type fakeRepository struct {
	create func(ctx context.Context, input CreateInput) (Booking, error)
}

func (f fakeRepository) Create(ctx context.Context, input CreateInput) (Booking, error) {
	if f.create == nil {
		return Booking{}, nil
	}

	return f.create(ctx, input)
}

type fakeRoomReader struct {
	getByID func(ctx context.Context, id uuid.UUID) (room.Room, error)
}

func (f fakeRoomReader) GetByID(ctx context.Context, id uuid.UUID) (room.Room, error) {
	return f.getByID(ctx, id)
}

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

func validInput(now time.Time) CreateInput {
	return CreateInput{
		RoomID:         uuid.New(),
		OrganizerID:    uuid.New(),
		Title:          "Team sync",
		StartsAt:       now.Add(time.Hour),
		EndsAt:         now.Add(2 * time.Hour),
		AttendeesCount: 5,
	}
}

func activeRoom(capacity int) room.Room {
	return room.Room{
		ID:       uuid.New(),
		Name:     "Alpha",
		Location: "HQ",
		Capacity: capacity,
		IsActive: true,
	}
}

func TestServiceCreate(t *testing.T) {
	now := time.Date(2026, 9, 10, 9, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		input     CreateInput
		room      room.Room
		roomErr   error
		wantErrIs error
	}{
		{
			name:  "valid booking",
			input: validInput(now),
			room:  activeRoom(10),
		},
		{
			name:      "room not found",
			input:     validInput(now),
			roomErr:   room.ErrNotFound,
			wantErrIs: ErrRoomNotFound,
		},
		{
			name:      "inactive room",
			input:     validInput(now),
			room:      room.Room{Capacity: 10, IsActive: false},
			wantErrIs: ErrRoomInactive,
		},
		{
			name: "capacity exceeded",
			input: func() CreateInput {
				in := validInput(now)
				in.AttendeesCount = 50
				return in
			}(),
			room:      activeRoom(10),
			wantErrIs: ErrCapacityExceeded,
		},
		{
			name: "starts in the past",
			input: func() CreateInput {
				in := validInput(now)
				in.StartsAt = now.Add(-time.Hour)
				in.EndsAt = now.Add(time.Hour)
				return in
			}(),
			room: activeRoom(10),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := fakeRepository{
				create: func(_ context.Context, input CreateInput) (Booking, error) {
					return Booking{
						ID:             uuid.New(),
						RoomID:         input.RoomID,
						OrganizerID:    input.OrganizerID,
						Title:          input.Title,
						StartsAt:       input.StartsAt,
						EndsAt:         input.EndsAt,
						Status:         StatusConfirmed,
						AttendeesCount: input.AttendeesCount,
					}, nil
				},
			}

			rooms := fakeRoomReader{
				getByID: func(_ context.Context, _ uuid.UUID) (room.Room, error) {
					return tt.room, tt.roomErr
				},
			}

			service := NewService(repository, rooms, fixedClock{now: now})

			_, err := service.Create(context.Background(), tt.input)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}
				return
			}

			if tt.name == "starts in the past" {
				var validationErr ValidationError
				if !errors.As(err, &validationErr) {
					t.Fatalf("expected validation error, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
