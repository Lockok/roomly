package booking

import "context"

type UseCase interface {
	Create(ctx context.Context, input CreateInput) (Booking, error)
}
