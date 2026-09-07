package booking

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/Lockok/roomly/internal/platform/optional"
)

type CreateRequest struct {
	RoomID         string  `json:"room_id"`
	OrganizerID    string  `json:"organizer_id"`
	Title          string  `json:"title"`
	Description    *string `json:"description"`
	StartsAt       string  `json:"starts_at"`
	EndsAt         string  `json:"ends_at"`
	AttendeesCount int     `json:"attendees_count"`
}

type BookingResponse struct {
	ID             string  `json:"id"`
	RoomID         string  `json:"room_id"`
	OrganizerID    string  `json:"organizer_id"`
	Title          string  `json:"title"`
	Description    *string `json:"description,omitempty"`
	StartsAt       string  `json:"starts_at"`
	EndsAt         string  `json:"ends_at"`
	Status         Status  `json:"status"`
	AttendeesCount int     `json:"attendees_count"`
	CancelledAt    *string `json:"cancelled_at,omitempty"`
	CancelledBy    *string `json:"cancelled_by,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

type UpdateRequest struct {
	Title          optional.Optional[string] `json:"title"`
	Description    optional.Optional[string] `json:"description"`
	StartsAt       optional.Optional[string] `json:"starts_at"`
	EndsAt         optional.Optional[string] `json:"ends_at"`
	AttendeesCount optional.Optional[int]    `json:"attendees_count"`
}

type CancelRequest struct {
	CancelledBy string `json:"cancelled_by"`
}

type ListResponse struct {
	Items []BookingResponse `json:"items"`
}

// toInput converts an UpdateRequest into a domain UpdateInput, parsing the
// RFC3339 timestamps for starts_at / ends_at when they are provided.
func (req UpdateRequest) toInput(id uuid.UUID) (UpdateInput, error) {
	input := UpdateInput{
		ID:             id,
		Title:          req.Title,
		Description:    req.Description,
		AttendeesCount: req.AttendeesCount,
	}

	if req.StartsAt.Set {
		input.StartsAt.Set = true
		if req.StartsAt.Value != nil {
			startsAt, err := time.Parse(time.RFC3339, *req.StartsAt.Value)
			if err != nil {
				return UpdateInput{}, fmt.Errorf("starts_at must be a valid RFC3339 timestamp")
			}
			input.StartsAt.Value = &startsAt
		}
	}

	if req.EndsAt.Set {
		input.EndsAt.Set = true
		if req.EndsAt.Value != nil {
			endsAt, err := time.Parse(time.RFC3339, *req.EndsAt.Value)
			if err != nil {
				return UpdateInput{}, fmt.Errorf("ends_at must be a valid RFC3339 timestamp")
			}
			input.EndsAt.Value = &endsAt
		}
	}

	return input, nil
}

// parseListFilter builds a ListFilter from query parameters:
// room_id, organizer_id, status, from, to.
func parseListFilter(r *http.Request) (ListFilter, error) {
	q := r.URL.Query()
	var filter ListFilter

	if v := q.Get("room_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return ListFilter{}, fmt.Errorf("room_id must be a valid UUID")
		}
		filter.RoomID = &id
	}

	if v := q.Get("organizer_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return ListFilter{}, fmt.Errorf("organizer_id must be a valid UUID")
		}
		filter.OrganizerID = &id
	}

	if v := q.Get("status"); v != "" {
		status := Status(v)
		if status != StatusConfirmed && status != StatusCancelled {
			return ListFilter{}, fmt.Errorf("status must be one of: confirmed, cancelled")
		}
		filter.Status = &status
	}

	if v := q.Get("from"); v != "" {
		from, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return ListFilter{}, fmt.Errorf("from must be a valid RFC3339 timestamp")
		}
		filter.From = &from
	}

	if v := q.Get("to"); v != "" {
		to, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return ListFilter{}, fmt.Errorf("to must be a valid RFC3339 timestamp")
		}
		filter.To = &to
	}

	return filter, nil
}

func NewListResponse(bookings []Booking) ListResponse {
	items := make([]BookingResponse, 0, len(bookings))
	for _, b := range bookings {
		items = append(items, NewBookingResponse(b))
	}

	return ListResponse{Items: items}
}

func NewBookingResponse(booking Booking) BookingResponse {
	response := BookingResponse{
		ID:             booking.ID.String(),
		RoomID:         booking.RoomID.String(),
		OrganizerID:    booking.OrganizerID.String(),
		Title:          booking.Title,
		Description:    booking.Description,
		StartsAt:       booking.StartsAt.Format(time.RFC3339),
		EndsAt:         booking.EndsAt.Format(time.RFC3339),
		Status:         booking.Status,
		AttendeesCount: booking.AttendeesCount,
		CreatedAt:      booking.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      booking.UpdatedAt.Format(time.RFC3339),
	}

	if booking.CancelledAt != nil {
		cancelledAt := booking.CancelledAt.Format(time.RFC3339)
		response.CancelledAt = &cancelledAt
	}

	if booking.CancelledBy != nil {
		cancelledBy := booking.CancelledBy.String()
		response.CancelledBy = &cancelledBy
	}

	return response
}
