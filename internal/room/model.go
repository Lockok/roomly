package room

import (
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID          uuid.UUID
	Name        string
	Location    string
	Floor       *int
	Capacity    int
	Equipment   []string
	Description *string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateInput struct {
	Name        string
	Location    string
	Floor       *int
	Capacity    int
	Equipment   []string
	Description *string
}
