package user

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, input CreateInput) (User, error)
	List(ctx context.Context, filter ListFilter) ([]User, error)
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
}
