package booking

import (
	"time"

	"github.com/Lockok/roomly/internal/platform/optional"
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

type ListFilter struct {
	RoomID      *uuid.UUID
	OrganizerID *uuid.UUID
	Status      *Status
	From        *time.Time
	To          *time.Time
}

type Actor struct {
	ID   uuid.UUID
	Role string
}

type UpdateInput struct {
	ID             uuid.UUID
	Actor          Actor
	Title          optional.Optional[string]
	Description    optional.Optional[string]
	StartsAt       optional.Optional[time.Time]
	EndsAt         optional.Optional[time.Time]
	AttendeesCount optional.Optional[int]
}

type CancelInput struct {
	ID          uuid.UUID
	CancelledBy uuid.UUID
	Actor       Actor
}
