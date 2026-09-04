package booking

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	minBookingDuration = 15 * time.Minute
	maxBookingDuration = 8 * time.Hour
)

func (input CreateInput) Validate(now time.Time) error {
	if input.RoomID == uuid.Nil {
		return ValidationError{Message: "room_id is required"}
	}

	if input.OrganizerID == uuid.Nil {
		return ValidationError{Message: "organizer_id is required"}
	}

	if strings.TrimSpace(input.Title) == "" {
		return ValidationError{Message: "title is required"}
	}

	if input.StartsAt.IsZero() {
		return ValidationError{Message: "starts_at is required"}
	}

	if input.EndsAt.IsZero() {
		return ValidationError{Message: "ends_at is required"}
	}

	if !input.EndsAt.After(input.StartsAt) {
		return ValidationError{Message: "ends_at must be after starts_at"}
	}

	if input.StartsAt.Before(now) {
		return ValidationError{Message: "starts_at must not be in the past"}
	}

	duration := input.EndsAt.Sub(input.StartsAt)
	if duration < minBookingDuration {
		return ValidationError{Message: "booking duration must be at least 15 minutes"}
	}

	if duration > maxBookingDuration {
		return ValidationError{Message: "booking duration must not exceed 8 hours"}
	}

	if input.AttendeesCount <= 0 {
		return ValidationError{Message: "attendees_count must be greater than zero"}
	}

	return nil
}
