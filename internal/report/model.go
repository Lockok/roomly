package report

import (
	"time"

	"github.com/google/uuid"
)

type RoomUsageInput struct {
	From time.Time
	To   time.Time
}

type RoomUsage struct {
	RoomID          uuid.UUID
	RoomName        string
	Location        string
	BookingsCount   int
	BookedMinutes   int
	UtilizationRate float64
}
