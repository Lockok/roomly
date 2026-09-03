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
	create  func(ctx context.Context, input CreateInput) (Room, error)
	list    func(ctx context.Context) ([]Room, error)
	getByID func(ctx context.Context, id uuid.UUID) (Room, error)
}

func (f fakeUseCase) Create(ctx context.Context, input CreateInput) (Room, error) {
	return f.create(ctx, input)
}

func (f fakeUseCase) List(ctx context.Context) ([]Room, error) {
	if f.list == nil {
		return nil, nil
	}

	return f.list(ctx)
}

func (f fakeUseCase) GetByID(ctx context.Context, id uuid.UUID) (Room, error) {
	if f.getByID == nil {
		return Room{}, ErrNotFound
	}
	return f.getByID(ctx, id)
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
		list: func(_ context.Context) ([]Room, error) {
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
		list: func(_ context.Context) ([]Room, error) {
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
