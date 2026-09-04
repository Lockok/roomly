package booking

import "context"

type Repository interface {
	Create(ctx context.Context, input CreateInput) (Booking, error)
}
