package room

import "context"

type UseCase interface {
	Create(ctx context.Context, input CreateInput) (Room, error)
	List(ctx context.Context) ([]Room, error)
}
