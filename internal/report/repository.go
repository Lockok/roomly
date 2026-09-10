package report

import "context"

type Repository interface {
	RoomUsage(ctx context.Context, input RoomUsageInput) ([]RoomUsage, error)
}
