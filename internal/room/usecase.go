package room

import (
	"context"

	"github.com/google/uuid"
)

type UseCase interface {
	Create(ctx context.Context, input CreateInput) (Room, error)
	List(ctx context.Context) ([]Room, error)
	GetByID(ctx context.Context, id uuid.UUID) (Room, error)
}
