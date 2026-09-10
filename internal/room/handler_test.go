package room

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeUseCase struct {
	create        func(ctx context.Context, input CreateInput) (Room, error)
	list          func(ctx context.Context, filter ListFilter) ([]Room, error)
	listAvailable func(ctx context.Context, input AvailabilityInput) ([]Room, error)
	getByID       func(ctx context.Context, id uuid.UUID) (Room, error)
	update        func(ctx context.Context, input UpdateInput) (Room, error)
}

func (f fakeUseCase) Create(ctx context.Context, input CreateInput) (Room, error) {
	return f.create(ctx, input)
}

func (f fakeUseCase) List(ctx context.Context, filter ListFilter) ([]Room, error) {
	if f.list == nil {
		return nil, nil
	}

	return f.list(ctx, filter)
}

func (f fakeUseCase) ListAvailable(ctx context.Context, input AvailabilityInput) ([]Room, error) {
	if f.listAvailable == nil {
		return nil, nil
	}

	return f.listAvailable(ctx, input)
}

func (f fakeUseCase) GetByID(ctx context.Context, id uuid.UUID) (Room, error) {
	if f.getByID == nil {
		return Room{}, ErrNotFound
	}
	return f.getByID(ctx, id)
}

func (f fakeUseCase) Update(ctx context.Context, input UpdateInput) (Room, error) {
	if f.update == nil {
		return Room{}, ErrNotFound
	}

	return f.update(ctx, input)
}

func TestHandlerCreateSuccess(t *testing.T) {
	expectedRoom := Room{
		ID:        uuid.New(),
		Name:      "Alpha",
		Location:  "HQ",
		Capacity:  10,
		Equipment: []string{"tv"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	useCase := fakeUseCase{
		create: func(_ context.Context, input CreateInput) (Room, error) {
			if input.Name != "Alpha" {
				t.Fatalf("expected name Alpha, got %q", input.Name)
			}

			return expectedRoom, nil
		},
	}

	handler := NewHandler(useCase)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/rooms",
		strings.NewReader(`{
			"name": "Alpha",
			"location": "HQ",
			"capacity": 10,
			"equipment": ["tv"]
			}`),
	)
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()

	handler.Create(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, response.Code)
	}

	var actualRoom RoomResponse

	if err := json.NewDecoder(response.Body).Decode(&actualRoom); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if actualRoom.ID != expectedRoom.ID.String() {
		t.Fatalf("expected room ID %s, got %s", expectedRoom.ID, actualRoom.ID)
	}

	if actualRoom.Name != expectedRoom.Name {
		t.Fatalf("expected room name %q, got %q", expectedRoom.Name, actualRoom.Name)
	}

	if actualRoom.Location != expectedRoom.Location {
		t.Fatalf("expected room location %q, got %q", expectedRoom.Location, actualRoom.Location)
	}
}

func TestHandlerListSuccess(t *testing.T) {
	useCase := fakeUseCase{
		list: func(_ context.Context, _ ListFilter) ([]Room, error) {
			return []Room{
				{
					ID:        uuid.New(),
					Name:      "Alpha",
					Location:  "HQ",
					Capacity:  10,
					Equipment: []string{"tv"},
					IsActive:  true,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			}, nil
		},
	}

	handler := NewHandler(useCase)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/rooms",
		nil,
	)
	response := httptest.NewRecorder()

	handler.List(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var rooms []RoomResponse
	if err := json.NewDecoder(response.Body).Decode(&rooms); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if len(rooms) != 1 {
		t.Fatalf("expected 1 room, got %d", len(rooms))
	}

	if rooms[0].Name != "Alpha" {
		t.Fatalf("expected room name %q, got %q", "Alpha", rooms[0].Name)
	}
}

func TestHandlerListEmpty(t *testing.T) {
	useCase := fakeUseCase{
		list: func(_ context.Context, _ ListFilter) ([]Room, error) {
			return []Room{}, nil
		},
	}

	handler := NewHandler(useCase)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/rooms",
		nil,
	)
	response := httptest.NewRecorder()

	handler.List(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	if response.Body.String() != "[]\n" {
		t.Fatalf("expected empty JSON array, got %q", response.Body.String())
	}
}

func TestHandlerGetByIDInvalidID(t *testing.T) {
	handler := NewHandler(fakeUseCase{})

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/rooms/not-a-uuid",
		nil,
	)
	request.SetPathValue("id", "not-a-uuid")

	response := httptest.NewRecorder()

	handler.GetByID(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}

	if response.Body.String() != "{\"code\":\"INVALID_ROOM_ID\",\"message\":\"room ID must be a valid UUID\"}\n" {
		t.Fatalf("unexpected response body: %q", response.Body.String())
	}
}

func TestHandlerGetByIDNotFound(t *testing.T) {
	handler := NewHandler(fakeUseCase{
		getByID: func(_ context.Context, _ uuid.UUID) (Room, error) {
			return Room{}, ErrNotFound
		},
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/rooms/00000000-0000-0000-0000-000000000000",
		nil,
	)
	request.SetPathValue("id", "00000000-0000-0000-0000-000000000000")

	response := httptest.NewRecorder()

	handler.GetByID(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}

	if response.Body.String() != "{\"code\":\"ROOM_NOT_FOUND\",\"message\":\"room not found\"}\n" {
		t.Fatalf("unexpected response body: %q", response.Body.String())
	}
}

func TestHandlerUpdateSuccess(t *testing.T) {
	roomID := uuid.New()

	useCase := fakeUseCase{
		update: func(_ context.Context, input UpdateInput) (Room, error) {
			if input.ID != roomID {
				t.Fatalf("expected room ID %s, got %s", roomID, input.ID)
			}

			if !input.Capacity.Set || input.Capacity.Value == nil || *input.Capacity.Value != 12 {
				t.Fatal("expected capacity update to 12")
			}

			if !input.Floor.Set || input.Floor.Value != nil {
				t.Fatal("expected floor to be explicitly set to null")
			}

			return Room{
				ID:        roomID,
				Name:      "Alpha",
				Location:  "HQ",
				Capacity:  12,
				Equipment: []string{"tv"},
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	handler := NewHandler(useCase)

	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/rooms/"+roomID.String(),
		strings.NewReader(`{
			"capacity": 12,
			"floor": null
		}`),
	)
	request.SetPathValue("id", roomID.String())

	response := httptest.NewRecorder()

	handler.Update(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
}

func TestHandlerUpdateInvalidID(t *testing.T) {
	handler := NewHandler(fakeUseCase{})

	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/rooms/not-a-uuid",
		strings.NewReader(`{"capacity": 12}`),
	)
	request.SetPathValue("id", "not-a-uuid")

	response := httptest.NewRecorder()

	handler.Update(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}

	if response.Body.String() != "{\"code\":\"INVALID_ROOM_ID\",\"message\":\"room ID must be a valid UUID\"}\n" {
		t.Fatalf("unexpected response body: %q", response.Body.String())
	}
}

func TestHandlerUpdateRejectsUnknownField(t *testing.T) {
	roomID := uuid.New()

	handler := NewHandler(fakeUseCase{
		update: func(_ context.Context, _ UpdateInput) (Room, error) {
			t.Fatal("use case must not be called for invalid JSON")
			return Room{}, nil
		},
	})

	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/rooms/"+roomID.String(),
		strings.NewReader(`{"unknown_field": "value"}`),
	)
	request.SetPathValue("id", roomID.String())

	response := httptest.NewRecorder()

	handler.Update(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}

	if response.Body.String() != "{\"code\":\"INVALID_JSON\",\"message\":\"request body must contain valid JSON\"}\n" {
		t.Fatalf("unexpected response body: %q", response.Body.String())
	}
}

func TestHandlerListWithActiveFilter(t *testing.T) {
	useCase := fakeUseCase{
		list: func(_ context.Context, filter ListFilter) ([]Room, error) {
			if filter.IsActive == nil {
				t.Fatal("expected active filter")
			}

			if !*filter.IsActive {
				t.Fatal("expected active filter to be true")
			}

			return []Room{}, nil
		},
	}

	handler := NewHandler(useCase)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/rooms?active=true",
		nil,
	)
	response := httptest.NewRecorder()

	handler.List(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
}

func TestHandlerListRejectsInvalidActiveFilter(t *testing.T) {
	handler := NewHandler(fakeUseCase{
		list: func(_ context.Context, _ ListFilter) ([]Room, error) {
			t.Fatal("use case must not be called for invalid filter")
			return nil, nil
		},
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/rooms?active=maybe",
		nil,
	)
	response := httptest.NewRecorder()

	handler.List(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}

	if response.Body.String() != "{\"code\":\"INVALID_ACTIVE_FILTER\",\"message\":\"active must be true or false\"}\n" {
		t.Fatalf("unexpected response body: %q", response.Body.String())
	}
}

func TestHandlerListAvailableSuccess(t *testing.T) {
	startsAt := "2026-09-10T10:00:00+03:00"
	endsAt := "2026-09-10T11:00:00+03:00"

	useCase := fakeUseCase{
		listAvailable: func(
			_ context.Context,
			input AvailabilityInput,
		) ([]Room, error) {
			if input.StartsAt.Format(time.RFC3339) != startsAt {
				t.Fatalf(
					"expected starts_at %s, got %s",
					startsAt,
					input.StartsAt.Format(time.RFC3339),
				)
			}

			if input.EndsAt.Format(time.RFC3339) != endsAt {
				t.Fatalf(
					"expected ends_at %s, got %s",
					endsAt,
					input.EndsAt.Format(time.RFC3339),
				)
			}

			return []Room{}, nil
		},
	}

	handler := NewHandler(useCase)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/rooms/available?starts_at=2026-09-10T10%3A00%3A00%2B03%3A00&ends_at=2026-09-10T11%3A00%3A00%2B03%3A00",
		nil,
	)
	response := httptest.NewRecorder()

	handler.ListAvailable(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	if response.Body.String() != "[]\n" {
		t.Fatalf("expected empty JSON array, got %q", response.Body.String())
	}
}

func TestHandlerListAvailableRejectsInvalidStartsAt(t *testing.T) {
	handler := NewHandler(fakeUseCase{
		listAvailable: func(
			_ context.Context,
			_ AvailabilityInput,
		) ([]Room, error) {
			t.Fatal("use case must not be called for invalid starts_at")
			return nil, nil
		},
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/rooms/available?starts_at=invalid&ends_at=2026-09-10T11%3A00%3A00%2B03%3A00",
		nil,
	)
	response := httptest.NewRecorder()

	handler.ListAvailable(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}

	if response.Body.String() != "{\"code\":\"INVALID_STARTS_AT\",\"message\":\"starts_at must be a valid RFC3339 timestamp\"}\n" {
		t.Fatalf("unexpected response body: %q", response.Body.String())
	}
}

func TestHandlerCreateRejectsMultipleJSONValues(t *testing.T) {
	handler := NewHandler(fakeUseCase{
		create: func(_ context.Context, _ CreateInput) (Room, error) {
			t.Fatal("use case must not be called for invalid JSON")
			return Room{}, nil
		},
	})

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/rooms",
		strings.NewReader(`{"name":"Alpha","location":"HQ","capacity":10} {"name":"Beta","location":"HQ","capacity":8}`),
	)
	response := httptest.NewRecorder()

	handler.Create(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}

	if response.Body.String() != "{\"code\":\"INVALID_JSON\",\"message\":\"request body must contain valid JSON\"}\n" {
		t.Fatalf("unexpected response body: %q", response.Body.String())
	}
}
func TestHandlerListAvailableWithFilters(t *testing.T) {
	startsAt := "2026-09-10T10:00:00Z"
	endsAt := "2026-09-10T11:00:00Z"

	useCase := fakeUseCase{
		listAvailable: func(_ context.Context, input AvailabilityInput) ([]Room, error) {
			if input.StartsAt.Format(time.RFC3339) != startsAt {
				t.Fatalf("unexpected starts_at: %s", input.StartsAt.Format(time.RFC3339))
			}

			if input.EndsAt.Format(time.RFC3339) != endsAt {
				t.Fatalf("unexpected ends_at: %s", input.EndsAt.Format(time.RFC3339))
			}

			if input.MinCapacity == nil || *input.MinCapacity != 8 {
				t.Fatal("expected min_capacity=8")
			}

			if input.Location == nil || *input.Location != "HQ" {
				t.Fatal("expected location=HQ")
			}

			if len(input.Equipment) != 2 ||
				input.Equipment[0] != "tv" ||
				input.Equipment[1] != "whiteboard" {
				t.Fatalf("unexpected equipment: %#v", input.Equipment)
			}

			return []Room{}, nil
		},
	}

	handler := NewHandler(useCase)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/rooms/available?starts_at="+startsAt+
			"&ends_at="+endsAt+
			"&min_capacity=8"+
			"&location=HQ"+
			"&equipment=tv"+
			"&equipment=whiteboard",
		nil,
	)
	response := httptest.NewRecorder()

	handler.ListAvailable(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
}
