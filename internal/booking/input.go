package booking

import (
	"time"

	"github.com/google/uuid"
)

type CreateInput struct {
	RoomID         uuid.UUID
	OrganizerID    uuid.UUID
	Title          string
	Description    *string
	StartsAt       time.Time
	EndsAt         time.Time
	AttendeesCount int
}
