package user

import (
	"context"

	"github.com/google/uuid"
)

type UseCase interface {
	Create(ctx context.Context, input CreateInput) (User, error)
	List(ctx context.Context, filter ListFilter) ([]User, error)
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	UpdateStatus(ctx context.Context, input UpdateStatusInput) (User, error)
}
