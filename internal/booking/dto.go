package booking

import "time"

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
