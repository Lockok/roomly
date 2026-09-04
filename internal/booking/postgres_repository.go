package booking

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const bookingNoOverlapConstraint = "bookings_no_overlapping_confirmed"

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool: pool,
	}
}

func (r *PostgresRepository) Create(ctx context.Context, input CreateInput) (Booking, error) {
	const query = `
		INSERT INTO bookings (
			room_id,
			organizer_id,
			title,
			description,
			starts_at,
			ends_at,
			attendees_count
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING
			id,
			room_id,
			organizer_id,
			title,
			description,
			starts_at,
			ends_at,
			status,
			attendees_count,
			cancelled_at,
			cancelled_by,
			created_at,
			updated_at;
	`

	var result Booking

	err := r.pool.QueryRow(
		ctx,
		query,
		input.RoomID,
		input.OrganizerID,
		input.Title,
		input.Description,
		input.StartsAt,
		input.EndsAt,
		input.AttendeesCount,
	).Scan(
		&result.ID,
		&result.RoomID,
		&result.OrganizerID,
		&result.Title,
		&result.Description,
		&result.StartsAt,
		&result.EndsAt,
		&result.Status,
		&result.AttendeesCount,
		&result.CancelledAt,
		&result.CancelledBy,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23P01" &&
			pgErr.ConstraintName == bookingNoOverlapConstraint {
			return Booking{}, ErrBookingConflict
		}

		if errors.Is(err, pgx.ErrNoRows) {
			return Booking{}, fmt.Errorf("create booking: no row returned")
		}

		return Booking{}, fmt.Errorf("create booking: %w", err)
	}

	return result, nil
}
