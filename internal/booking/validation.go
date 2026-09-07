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

	if err := validateTimeRange(input.StartsAt, input.EndsAt, now); err != nil {
		return err
	}

	if input.AttendeesCount <= 0 {
		return ValidationError{Message: "attendees_count must be greater than zero"}
	}

	return nil
}

// validateTimeRange checks that a booking time range is valid: ends after
// starts, not in the past, and within the allowed duration bounds.
func validateTimeRange(startsAt, endsAt, now time.Time) error {
	if !endsAt.After(startsAt) {
		return ValidationError{Message: "ends_at must be after starts_at"}
	}

	if startsAt.Before(now) {
		return ValidationError{Message: "starts_at must not be in the past"}
	}

	duration := endsAt.Sub(startsAt)
	if duration < minBookingDuration {
		return ValidationError{Message: "booking duration must be at least 15 minutes"}
	}

	if duration > maxBookingDuration {
		return ValidationError{Message: "booking duration must not exceed 8 hours"}
	}

	return nil
}

// Validate checks the fields provided in an update. Only fields that are set
// are validated. When both starts_at and ends_at are provided together, the
// full range is validated here; the mixed case (only one bound changed) is
// validated in the service after loading the current booking.
func (input UpdateInput) Validate(now time.Time) error {
	if input.Title.Set {
		if input.Title.Value == nil || strings.TrimSpace(*input.Title.Value) == "" {
			return ValidationError{Message: "title must not be empty"}
		}
	}

	if input.AttendeesCount.Set {
		if input.AttendeesCount.Value == nil || *input.AttendeesCount.Value <= 0 {
			return ValidationError{Message: "attendees_count must be greater than zero"}
		}
	}

	if input.StartsAt.Set && input.StartsAt.Value == nil {
		return ValidationError{Message: "starts_at must not be null"}
	}

	if input.EndsAt.Set && input.EndsAt.Value == nil {
		return ValidationError{Message: "ends_at must not be null"}
	}

	if input.StartsAt.Set && input.EndsAt.Set {
		if err := validateTimeRange(*input.StartsAt.Value, *input.EndsAt.Value, now); err != nil {
			return err
		}
	}

	return nil
}

func (f ListFilter) Validate() error {
	if f.From != nil && f.To != nil && !f.To.After(*f.From) {
		return ValidationError{Message: "to must be after from"}
	}

	return nil
}
