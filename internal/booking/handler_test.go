package booking

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeUseCase struct {
	list func(ctx context.Context, filter ListFilter) ([]Booking, error)
}

func (f fakeUseCase) Create(context.Context, CreateInput) (Booking, error) {
	return Booking{}, nil
}

func (f fakeUseCase) List(ctx context.Context, filter ListFilter) ([]Booking, error) {
	if f.list == nil {
		return nil, nil
	}

	return f.list(ctx, filter)
}

func (f fakeUseCase) GetByID(context.Context, uuid.UUID) (Booking, error) {
	return Booking{}, ErrNotFound
}

func (f fakeUseCase) Update(context.Context, UpdateInput) (Booking, error) {
	return Booking{}, nil
}

func (f fakeUseCase) Cancel(context.Context, CancelInput) (Booking, error) {
	return Booking{}, nil
}

func TestHandlerRoomCalendar(t *testing.T) {
	roomID := uuid.New()
	from := "2026-09-10T10:00:00Z"
	to := "2026-09-10T11:00:00Z"

	useCase := fakeUseCase{
		list: func(_ context.Context, filter ListFilter) ([]Booking, error) {
			if filter.RoomID == nil || *filter.RoomID != roomID {
				t.Fatal("expected room ID filter")
			}

			if filter.Status == nil || *filter.Status != StatusConfirmed {
				t.Fatal("expected confirmed status filter")
			}

			if filter.From == nil || filter.From.Format(time.RFC3339) != from {
				t.Fatal("unexpected from filter")
			}

			if filter.To == nil || filter.To.Format(time.RFC3339) != to {
				t.Fatal("unexpected to filter")
			}

			return []Booking{}, nil
		},
	}

	handler := NewHandler(useCase)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/rooms/"+roomID.String()+"/calendar?from="+from+"&to="+to,
		nil,
	)
	request.SetPathValue("id", roomID.String())

	response := httptest.NewRecorder()

	handler.RoomCalendar(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
}
