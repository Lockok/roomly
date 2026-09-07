package booking

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const bookingColumns = `
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
	updated_at
`

const bookingNoOverlapConstraint = "bookings_no_overlapping_confirmed"

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool: pool,
	}
}

func scanBooking(row pgx.Row) (Booking, error) {
	var b Booking

	err := row.Scan(
		&b.ID,
		&b.RoomID,
		&b.OrganizerID,
		&b.Title,
		&b.Description,
		&b.StartsAt,
		&b.EndsAt,
		&b.Status,
		&b.AttendeesCount,
		&b.CancelledAt,
		&b.CancelledBy,
		&b.CreatedAt,
		&b.UpdatedAt,
	)

	return b, err
}

func isOverlapConflict(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) &&
		pgErr.Code == "23P01" &&
		pgErr.ConstraintName == bookingNoOverlapConstraint
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
		RETURNING ` + bookingColumns + `;`

	row := r.pool.QueryRow(
		ctx,
		query,
		input.RoomID,
		input.OrganizerID,
		input.Title,
		input.Description,
		input.StartsAt,
		input.EndsAt,
		input.AttendeesCount,
	)

	result, err := scanBooking(row)
	if err != nil {
		if isOverlapConflict(err) {
			return Booking{}, ErrBookingConflict
		}

		if errors.Is(err, pgx.ErrNoRows) {
			return Booking{}, fmt.Errorf("create booking: no row returned")
		}

		return Booking{}, fmt.Errorf("create booking: %w", err)
	}

	return result, nil
}

func (r *PostgresRepository) List(ctx context.Context, filter ListFilter) ([]Booking, error) {
	const query = `
		SELECT ` + bookingColumns + `
		FROM bookings
		WHERE ($1::uuid IS NULL OR room_id = $1)
		  AND ($2::uuid IS NULL OR organizer_id = $2)
		  AND ($3::booking_status IS NULL OR status = $3)
		  AND ($4::timestamptz IS NULL OR ends_at > $4)
		  AND ($5::timestamptz IS NULL OR starts_at < $5)
		ORDER BY starts_at ASC;
	`

	rows, err := r.pool.Query(ctx, query, filter.RoomID, filter.OrganizerID, filter.Status, filter.From, filter.To)
	if err != nil {
		return nil, fmt.Errorf("list bookings: %w", err)
	}
	defer rows.Close()

	bookings := make([]Booking, 0)

	for rows.Next() {
		result, err := scanBooking(rows)
		if err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}

		bookings = append(bookings, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bookings: %w", err)
	}

	return bookings, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (Booking, error) {
	const query = `
		SELECT ` + bookingColumns + `
		FROM bookings
		WHERE id = $1;
	`

	result, err := scanBooking(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Booking{}, ErrNotFound
		}

		return Booking{}, fmt.Errorf("get booking by ID: %w", err)
	}

	return result, nil
}

func (r *PostgresRepository) Update(ctx context.Context, input UpdateInput) (Booking, error) {
	const query = `
		UPDATE bookings
		SET
			title = CASE WHEN $2 THEN $3 ELSE title END,
			description = CASE WHEN $4 THEN $5 ELSE description END,
			starts_at = CASE WHEN $6 THEN $7 ELSE starts_at END,
			ends_at = CASE WHEN $8 THEN $9 ELSE ends_at END,
			attendees_count = CASE WHEN $10 THEN $11 ELSE attendees_count END
		WHERE id = $1
		RETURNING ` + bookingColumns + `;`

	row := r.pool.QueryRow(
		ctx,
		query,
		input.ID,
		input.Title.Set,
		input.Title.Value,
		input.Description.Set,
		input.Description.Value,
		input.StartsAt.Set,
		input.StartsAt.Value,
		input.EndsAt.Set,
		input.EndsAt.Value,
		input.AttendeesCount.Set,
		input.AttendeesCount.Value,
	)

	result, err := scanBooking(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Booking{}, ErrNotFound
		}

		if isOverlapConflict(err) {
			return Booking{}, ErrBookingConflict
		}

		return Booking{}, fmt.Errorf("update booking: %w", err)
	}

	return result, nil
}

func (r *PostgresRepository) Cancel(ctx context.Context, input CancelInput) (Booking, error) {
	const query = `
		UPDATE bookings
		SET
			status = 'cancelled',
			cancelled_at = NOW(),
			cancelled_by = $2
		WHERE id = $1 AND status = 'confirmed'
		RETURNING ` + bookingColumns + `;`

	result, err := scanBooking(r.pool.QueryRow(ctx, query, input.ID, input.CancelledBy))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Booking{}, ErrNotFound
		}

		return Booking{}, fmt.Errorf("cancel booking: %w", err)
	}

	return result, nil
}
