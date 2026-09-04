package booking

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/Lockok/roomly/internal/platform/httputil"
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

	if err := httputil.DecodeJSON(w, r, &request); err != nil {
		httputil.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"request body must contain valid JSON",
		)
		return
	}

	roomID, err := uuid.Parse(request.RoomID)
	if err != nil {
		httputil.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_ROOM_ID",
			"room_id must be a valid UUID",
		)
		return
	}

	organizerID, err := uuid.Parse(request.OrganizerID)
	if err != nil {
		httputil.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_ORGANIZER_ID",
			"organizer_id must be a valid UUID",
		)
		return
	}

	startsAt, err := time.Parse(time.RFC3339, request.StartsAt)
	if err != nil {
		httputil.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_STARTS_AT",
			"starts_at must be a valid RFC3339 timestamp",
		)
		return
	}

	endsAt, err := time.Parse(time.RFC3339, request.EndsAt)
	if err != nil {
		httputil.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_ENDS_AT",
			"ends_at must be a valid RFC3339 timestamp",
		)
		return
	}

	input := CreateInput{
		RoomID:         roomID,
		OrganizerID:    organizerID,
		Title:          request.Title,
		Description:    request.Description,
		StartsAt:       startsAt,
		EndsAt:         endsAt,
		AttendeesCount: request.AttendeesCount,
	}

	createdBooking, err := h.useCase.Create(r.Context(), input)
	if err != nil {
		writeCreateError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, NewBookingResponse(createdBooking))
}

func writeCreateError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrRoomNotFound):
		httputil.WriteError(w, http.StatusNotFound, "ROOM_NOT_FOUND", "room not found")
	case errors.Is(err, ErrRoomInactive):
		httputil.WriteError(w, http.StatusConflict, "ROOM_INACTIVE", "room is not active")
	case errors.Is(err, ErrCapacityExceeded):
		httputil.WriteError(w, http.StatusConflict, "CAPACITY_EXCEEDED", "attendees exceed room capacity")
	case errors.Is(err, ErrBookingConflict):
		httputil.WriteError(w, http.StatusConflict, "BOOKING_CONFLICT", "room is already booked for this time range")
	default:
		var validationErr ValidationError
		if errors.As(err, &validationErr) {
			httputil.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", validationErr.Message)
			return
		}

		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}
