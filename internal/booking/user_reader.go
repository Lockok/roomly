package booking

import (
	"context"

	"github.com/Lockok/roomly/internal/user"
	"github.com/google/uuid"
)

type UserReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (user.User, error)
}
