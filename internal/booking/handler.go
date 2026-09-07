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
		writeBookingError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, NewBookingResponse(createdBooking))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	filter, err := parseListFilter(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_QUERY", err.Error())
		return
	}

	bookings, err := h.useCase.List(r.Context(), filter)
	if err != nil {
		writeBookingError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, NewListResponse(bookings))
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_ID", "id must be a valid UUID")
		return
	}

	found, err := h.useCase.GetByID(r.Context(), id)
	if err != nil {
		writeBookingError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, NewBookingResponse(found))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_ID", "id must be a valid UUID")
		return
	}

	var request UpdateRequest

	if err := httputil.DecodeJSON(w, r, &request); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "request body must contain valid JSON")
		return
	}

	input, err := request.toInput(id)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	updated, err := h.useCase.Update(r.Context(), input)
	if err != nil {
		writeBookingError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, NewBookingResponse(updated))
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_ID", "id must be a valid UUID")
		return
	}

	var request CancelRequest

	if err := httputil.DecodeJSON(w, r, &request); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "request body must contain valid JSON")
		return
	}

	cancelledBy, err := uuid.Parse(request.CancelledBy)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_CANCELLED_BY", "cancelled_by must be a valid UUID")
		return
	}

	cancelled, err := h.useCase.Cancel(r.Context(), CancelInput{ID: id, CancelledBy: cancelledBy})
	if err != nil {
		writeBookingError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, NewBookingResponse(cancelled))
}

func writeBookingError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrRoomNotFound):
		httputil.WriteError(w, http.StatusNotFound, "ROOM_NOT_FOUND", "room not found")
	case errors.Is(err, ErrNotFound):
		httputil.WriteError(w, http.StatusNotFound, "BOOKING_NOT_FOUND", "booking not found")
	case errors.Is(err, ErrRoomInactive):
		httputil.WriteError(w, http.StatusConflict, "ROOM_INACTIVE", "room is not active")
	case errors.Is(err, ErrCapacityExceeded):
		httputil.WriteError(w, http.StatusConflict, "CAPACITY_EXCEEDED", "attendees exceed room capacity")
	case errors.Is(err, ErrBookingConflict):
		httputil.WriteError(w, http.StatusConflict, "BOOKING_CONFLICT", "room is already booked for this time range")
	case errors.Is(err, ErrAlreadyCancelled):
		httputil.WriteError(w, http.StatusConflict, "ALREADY_CANCELLED", "booking is already cancelled")
	case errors.Is(err, ErrBookingStarted):
		httputil.WriteError(w, http.StatusConflict, "BOOKING_STARTED", "booking has already started or finished")
	default:
		var validationErr ValidationError
		if errors.As(err, &validationErr) {
			httputil.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", validationErr.Message)
			return
		}

		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}
