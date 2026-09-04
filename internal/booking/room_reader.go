package booking

import (
	"context"

	"github.com/Lockok/roomly/internal/room"
	"github.com/google/uuid"
)

type RoomReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (room.Room, error)
}
