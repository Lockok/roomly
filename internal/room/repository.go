package room

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, input CreateInput) (Room, error)
	List(ctx context.Context) ([]Room, error)
	GetByID(ctx context.Context, id uuid.UUID) (Room, error)
	Update(ctx context.Context, input UpdateInput) (Room, error)
}
