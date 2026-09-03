package room

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Lockok/roomly/internal/platform/httputil"
	"github.com/google/uuid"
)

type Handler struct {
	useCase UseCase
}

func NewHandler(useCase UseCase) *Handler {
	return &Handler{
		useCase: useCase,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request CreateRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		httputil.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"request body must containt valid JSON",
		)
		return
	}

	input := CreateInput{
		Name:        request.Name,
		Location:    request.Location,
		Floor:       request.Floor,
		Capacity:    request.Capacity,
		Equipment:   request.Equipment,
		Description: request.Description,
	}

	createdRoom, err := h.useCase.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			httputil.WriteError(
				w,
				http.StatusConflict,
				"ROOM_ALREADY_EXISTS",
				"room with this name already exists in this location",
			)
			return
		}

		var validationErr ValidationError
		if errors.As(err, &validationErr) {
			httputil.WriteError(
				w,
				http.StatusBadRequest,
				"VALIDATION_ERROR",
				validationErr.Message,
			)
			return
		}

		httputil.WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"internal server error",
		)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, NewRoomResponse(createdRoom))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	filter := ListFilter{}

	activeValue := r.URL.Query().Get("active")
	if activeValue != "" {
		isActive, err := strconv.ParseBool(activeValue)
		if err != nil {
			httputil.WriteError(
				w,
				http.StatusBadRequest,
				"INVALID_ACTIVE_FILTER",
				"active must be true or false",
			)
			return
		}

		filter.IsActive = &isActive
	}
	rooms, err := h.useCase.List(r.Context(), filter)

	if err != nil {
		httputil.WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"internal server error",
		)
		return
	}

	response := make([]RoomResponse, 0, len(rooms))
	for _, item := range rooms {
		response = append(response, NewRoomResponse(item))
	}

	httputil.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) ListAvailable(w http.ResponseWriter, r *http.Request) {
	startsAt, err := time.Parse(time.RFC3339, r.URL.Query().Get("starts_at"))
	if err != nil {
		httputil.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_STARTS_AT",
			"starts_at must be a valid RFC3339 timestamp",
		)
		return
	}

	endsAt, err := time.Parse(time.RFC3339, r.URL.Query().Get("ends_at"))
	if err != nil {
		httputil.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_ENDS_AT",
			"ends_at must be a valid RFC3339 timestamp",
		)
		return
	}

	rooms, err := h.useCase.ListAvailable(r.Context(), AvailabilityInput{
		StartsAt: startsAt,
		EndsAt:   endsAt,
	})
	if err != nil {
		var validationErr ValidationError
		if errors.As(err, &validationErr) {
			httputil.WriteError(
				w,
				http.StatusBadRequest,
				"VALIDATION_ERROR",
				validationErr.Message,
			)
			return
		}

		httputil.WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"internal server error",
		)
		return
	}

	response := make([]RoomResponse, 0, len(rooms))
	for _, item := range rooms {
		response = append(response, NewRoomResponse(item))
	}

	httputil.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httputil.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_ROOM_ID",
			"room ID must be a valid UUID",
		)
		return
	}

	foundRoom, err := h.useCase.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.WriteError(
				w,
				http.StatusNotFound,
				"ROOM_NOT_FOUND",
				"room not found",
			)
			return
		}

		httputil.WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"internal server error",
		)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, NewRoomResponse(foundRoom))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httputil.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_ROOM_ID",
			"room ID must be a valid UUID",
		)
		return
	}

	var request UpdateRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		httputil.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"request body must contain valid JSON",
		)
		return
	}

	input := UpdateInput{
		ID:          id,
		Name:        request.Name,
		Location:    request.Location,
		Floor:       request.Floor,
		Capacity:    request.Capacity,
		Equipment:   request.Equipment,
		Description: request.Description,
		IsActive:    request.IsActive,
	}

	updatedRoom, err := h.useCase.Update(r.Context(), input)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.WriteError(
				w,
				http.StatusNotFound,
				"ROOM_NOT_FOUND",
				"room not found",
			)
			return
		}

		if errors.Is(err, ErrAlreadyExists) {
			httputil.WriteError(
				w,
				http.StatusConflict,
				"ROOM_ALREADY_EXISTS",
				"room with this name already exists in this location",
			)
			return
		}

		var validationErr ValidationError
		if errors.As(err, &validationErr) {
			httputil.WriteError(
				w,
				http.StatusBadRequest,
				"VALIDATION_ERROR",
				validationErr.Message,
			)
			return
		}

		httputil.WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"internal server error",
		)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, NewRoomResponse(updatedRoom))
}
