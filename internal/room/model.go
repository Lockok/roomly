package room

import (
	"time"

	"github.com/Lockok/roomly/internal/platform/optional"
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

type UpdateInput struct {
	ID          uuid.UUID
	Name        optional.Optional[string]
	Location    optional.Optional[string]
	Floor       optional.Optional[int]
	Capacity    optional.Optional[int]
	Equipment   optional.Optional[[]string]
	Description optional.Optional[string]
	IsActive    optional.Optional[bool]
}

type ListFilter struct {
	IsActive *bool
}
