package room

import "context"

type Repository interface {
	Create(ctx context.Context, input CreateInput) (Room, error)
	List(ctx context.Context) ([]Room, error)
}
