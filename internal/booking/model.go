package booking

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusConfirmed Status = "confirmed"
	StatusCancelled Status = "cancelled"
)

type Booking struct {
	ID             uuid.UUID
	RoomID         uuid.UUID
	OrganizerID    uuid.UUID
	Title          string
	Description    *string
	StartsAt       time.Time
	EndsAt         time.Time
	Status         Status
	AttendeesCount int
	CancelledAt    *time.Time
	CancelledBy    *uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
