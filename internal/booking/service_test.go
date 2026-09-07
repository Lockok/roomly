package booking

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Lockok/roomly/internal/platform/optional"
	"github.com/Lockok/roomly/internal/room"
	"github.com/google/uuid"
)

type fakeRepository struct {
	create  func(ctx context.Context, input CreateInput) (Booking, error)
	list    func(ctx context.Context, filter ListFilter) ([]Booking, error)
	getByID func(ctx context.Context, id uuid.UUID) (Booking, error)
	update  func(ctx context.Context, input UpdateInput) (Booking, error)
	cancel  func(ctx context.Context, input CancelInput) (Booking, error)
}

func (f fakeRepository) Create(ctx context.Context, input CreateInput) (Booking, error) {
	if f.create == nil {
		return Booking{}, nil
	}

	return f.create(ctx, input)
}

func (f fakeRepository) List(ctx context.Context, filter ListFilter) ([]Booking, error) {
	if f.list == nil {
		return nil, nil
	}

	return f.list(ctx, filter)
}

func (f fakeRepository) GetByID(ctx context.Context, id uuid.UUID) (Booking, error) {
	if f.getByID == nil {
		return Booking{}, nil
	}

	return f.getByID(ctx, id)
}

func (f fakeRepository) Update(ctx context.Context, input UpdateInput) (Booking, error) {
	if f.update == nil {
		return Booking{}, nil
	}

	return f.update(ctx, input)
}

func (f fakeRepository) Cancel(ctx context.Context, input CancelInput) (Booking, error) {
	if f.cancel == nil {
		return Booking{}, nil
	}

	return f.cancel(ctx, input)
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

func confirmedBooking(now time.Time, roomID uuid.UUID) Booking {
	return Booking{
		ID:             uuid.New(),
		RoomID:         roomID,
		OrganizerID:    uuid.New(),
		Title:          "Team sync",
		StartsAt:       now.Add(time.Hour),
		EndsAt:         now.Add(2 * time.Hour),
		Status:         StatusConfirmed,
		AttendeesCount: 5,
	}
}

func TestServiceUpdate(t *testing.T) {
	now := time.Date(2026, 9, 10, 9, 0, 0, 0, time.UTC)
	roomID := uuid.New()

	newTitle := "Renamed"
	tooManyAttendees := 50
	pastStart := now.Add(-time.Hour)

	tests := []struct {
		name       string
		current    Booking
		currentErr error
		updateErr  error
		input      UpdateInput
		wantErrIs  error
	}{
		{
			name:    "update title",
			current: confirmedBooking(now, roomID),
			input: UpdateInput{
				Title: optional.Optional[string]{Set: true, Value: &newTitle},
			},
		},
		{
			name:       "booking not found",
			currentErr: ErrNotFound,
			input:      UpdateInput{Title: optional.Optional[string]{Set: true, Value: &newTitle}},
			wantErrIs:  ErrNotFound,
		},
		{
			name: "cannot update cancelled",
			current: func() Booking {
				b := confirmedBooking(now, roomID)
				b.Status = StatusCancelled
				return b
			}(),
			input:     UpdateInput{Title: optional.Optional[string]{Set: true, Value: &newTitle}},
			wantErrIs: ErrAlreadyCancelled,
		},
		{
			name: "cannot update started booking",
			current: func() Booking {
				b := confirmedBooking(now, roomID)
				b.StartsAt = now.Add(-time.Minute)
				return b
			}(),
			input:     UpdateInput{Title: optional.Optional[string]{Set: true, Value: &newTitle}},
			wantErrIs: ErrBookingStarted,
		},
		{
			name:    "attendees exceed capacity",
			current: confirmedBooking(now, roomID),
			input: UpdateInput{
				AttendeesCount: optional.Optional[int]{Set: true, Value: &tooManyAttendees},
			},
			wantErrIs: ErrCapacityExceeded,
		},
		{
			name:    "invalid partial time range",
			current: confirmedBooking(now, roomID),
			input: UpdateInput{
				StartsAt: optional.Optional[time.Time]{Set: true, Value: &pastStart},
			},
			wantErrIs: nil, // validation error, checked separately
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.ID = tt.current.ID

			repository := fakeRepository{
				getByID: func(_ context.Context, _ uuid.UUID) (Booking, error) {
					return tt.current, tt.currentErr
				},
				update: func(_ context.Context, input UpdateInput) (Booking, error) {
					result := tt.current
					if input.Title.Set && input.Title.Value != nil {
						result.Title = *input.Title.Value
					}
					return result, tt.updateErr
				},
			}

			rooms := fakeRoomReader{
				getByID: func(_ context.Context, _ uuid.UUID) (room.Room, error) {
					return activeRoom(10), nil
				},
			}

			service := NewService(repository, rooms, fixedClock{now: now})

			_, err := service.Update(context.Background(), tt.input)

			if tt.name == "invalid partial time range" {
				var validationErr ValidationError
				if !errors.As(err, &validationErr) {
					t.Fatalf("expected validation error, got %v", err)
				}
				return
			}

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestServiceCancel(t *testing.T) {
	now := time.Date(2026, 9, 10, 9, 0, 0, 0, time.UTC)
	roomID := uuid.New()

	t.Run("cancel confirmed booking", func(t *testing.T) {
		current := confirmedBooking(now, roomID)
		repository := fakeRepository{
			getByID: func(_ context.Context, _ uuid.UUID) (Booking, error) {
				return current, nil
			},
			cancel: func(_ context.Context, _ CancelInput) (Booking, error) {
				current.Status = StatusCancelled
				return current, nil
			},
		}

		service := NewService(repository, fakeRoomReader{}, fixedClock{now: now})

		result, err := service.Cancel(context.Background(), CancelInput{ID: current.ID, CancelledBy: uuid.New()})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result.Status != StatusCancelled {
			t.Fatalf("expected status cancelled, got %v", result.Status)
		}
	})

	t.Run("cannot cancel already cancelled", func(t *testing.T) {
		current := confirmedBooking(now, roomID)
		current.Status = StatusCancelled

		repository := fakeRepository{
			getByID: func(_ context.Context, _ uuid.UUID) (Booking, error) {
				return current, nil
			},
		}

		service := NewService(repository, fakeRoomReader{}, fixedClock{now: now})

		_, err := service.Cancel(context.Background(), CancelInput{ID: current.ID, CancelledBy: uuid.New()})
		if !errors.Is(err, ErrAlreadyCancelled) {
			t.Fatalf("expected ErrAlreadyCancelled, got %v", err)
		}
	})

	t.Run("booking not found", func(t *testing.T) {
		repository := fakeRepository{
			getByID: func(_ context.Context, _ uuid.UUID) (Booking, error) {
				return Booking{}, ErrNotFound
			},
		}

		service := NewService(repository, fakeRoomReader{}, fixedClock{now: now})

		_, err := service.Cancel(context.Background(), CancelInput{ID: uuid.New(), CancelledBy: uuid.New()})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}
