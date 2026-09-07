package booking

import (
	"context"

	"github.com/google/uuid"
)

type UseCase interface {
	Create(ctx context.Context, input CreateInput) (Booking, error)
	List(ctx context.Context, filter ListFilter) ([]Booking, error)
	GetByID(ctx context.Context, id uuid.UUID) (Booking, error)
	Update(ctx context.Context, input UpdateInput) (Booking, error)
	Cancel(ctx context.Context, input CancelInput) (Booking, error)
}
