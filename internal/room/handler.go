package room

import (
	"encoding/json"
	"errors"
	"net/http"

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
	rooms, err := h.useCase.List(r.Context())

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
