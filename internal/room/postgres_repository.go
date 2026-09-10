package room

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const roomsUniqueNamePerLocationConstraint = "rooms_unique_name_per_location"

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool: pool,
	}
}

func (r *PostgresRepository) Create(ctx context.Context, input CreateInput) (Room, error) {
	equipmentJSON, err := json.Marshal(input.Equipment)
	if err != nil {
		return Room{}, fmt.Errorf("marshal equipment: %w", err)
	}

	const query = `
		INSERT INTO rooms(
			name, 
			location,
			floor,
			capacity,
			equipment,
			description
		)
		VALUES($1, $2, $3, $4, $5::jsonb, $6) 
		RETURNING
			id,
			name,
			location,
			floor,
			capacity,
			equipment,
			description,
			is_active,
			created_at,
			updated_at;
	`

	var result Room

	err = r.pool.QueryRow(ctx, query, input.Name, input.Location, input.Floor, input.Capacity, equipmentJSON, input.Description).Scan(
		&result.ID,
		&result.Name,
		&result.Location,
		&result.Floor,
		&result.Capacity,
		&result.Equipment,
		&result.Description,
		&result.IsActive,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == roomsUniqueNamePerLocationConstraint {
			return Room{}, ErrAlreadyExists
		}

		if errors.Is(err, pgx.ErrNoRows) {
			return Room{}, fmt.Errorf("create room: no row returned")
		}

		return Room{}, fmt.Errorf("create room: %w", err)
	}

	return result, nil
}

func (r *PostgresRepository) List(ctx context.Context, filter ListFilter) ([]Room, error) {
	const query = `
		SELECT
			id,
			name,
			location,
			floor,
			capacity,
			equipment,
			description,
			is_active,
			created_at,
			updated_at
		FROM rooms
		WHERE ($1::boolean IS NULL OR is_active = $1)
		ORDER BY location ASC, name ASC;
	`
	rows, err := r.pool.Query(ctx, query, filter.IsActive)
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	defer rows.Close()

	rooms := make([]Room, 0)

	for rows.Next() {
		var result Room

		if err := rows.Scan(
			&result.ID,
			&result.Name,
			&result.Location,
			&result.Floor,
			&result.Capacity,
			&result.Equipment,
			&result.Description,
			&result.IsActive,
			&result.CreatedAt,
			&result.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan room: %w", err)
		}

		rooms = append(rooms, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rooms: %w", err)
	}

	return rooms, nil
}

func (r *PostgresRepository) ListAvailable(ctx context.Context, input AvailabilityInput) ([]Room, error) {
	const query = `
	SELECT
		r.id,
		r.name,
		r.location,
		r.floor,
		r.capacity,
		r.equipment,
		r.description,
		r.is_active,
		r.created_at,
		r.updated_at
	FROM rooms AS r
	WHERE r.is_active = TRUE
	  AND ($3::integer IS NULL OR r.capacity >= $3)
	  AND ($4::text IS NULL OR r.location = $4)
	  AND (
		COALESCE(array_length($5::text[], 1), 0) = 0
		OR r.equipment @> to_jsonb($5::text[])
	  )
	  AND NOT EXISTS (
		  SELECT 1
		  FROM bookings AS b
		  WHERE b.room_id = r.id
		    AND b.status = 'confirmed'
		    AND tstzrange(b.starts_at, b.ends_at, '[)')
		        && tstzrange($1, $2, '[)')
	  )
	ORDER BY r.location ASC, r.name ASC;
	`

	rows, err := r.pool.Query(ctx, query, input.StartsAt, input.EndsAt, input.MinCapacity, input.Location, input.Equipment)
	if err != nil {
		return nil, fmt.Errorf("list available error: %w", err)
	}
	defer rows.Close()

	rooms := make([]Room, 0)

	for rows.Next() {
		var result Room

		if err := rows.Scan(
			&result.ID,
			&result.Name,
			&result.Location,
			&result.Floor,
			&result.Capacity,
			&result.Equipment,
			&result.Description,
			&result.IsActive,
			&result.CreatedAt,
			&result.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan available room: %w", err)
		}

		rooms = append(rooms, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate available room: %w", err)
	}

	return rooms, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (Room, error) {
	const query = `
		SELECT
			id,
			name,
			location,
			floor,
			capacity,
			equipment,
			description,
			is_active,
			created_at,
			updated_at
		FROM rooms
		WHERE id = $1;
	`

	var result Room

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&result.ID,
		&result.Name,
		&result.Location,
		&result.Floor,
		&result.Capacity,
		&result.Equipment,
		&result.Description,
		&result.IsActive,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Room{}, ErrNotFound
		}

		return Room{}, fmt.Errorf("get room by ID: %w", err)
	}

	return result, nil
}

func (r *PostgresRepository) Update(ctx context.Context, input UpdateInput) (Room, error) {
	var equipmentJSON []byte
	var err error

	if input.Equipment.Set {
		equipmentJSON, err = json.Marshal(input.Equipment.Value)
		if err != nil {
			return Room{}, fmt.Errorf("marshal equipment: %w", err)
		}
	}

	const query = `
		UPDATE rooms
		SET
			name = CASE WHEN $2 THEN $3 ELSE name END,
			location = CASE WHEN $4 THEN $5 ELSE location END,
			floor = CASE WHEN $6 THEN $7 ELSE floor END,
			capacity = CASE WHEN $8 THEN $9 ELSE capacity END,
			equipment = CASE WHEN $10 THEN $11::jsonb ELSE equipment END,
			description = CASE WHEN $12 THEN $13 ELSE description END,
			is_active = CASE WHEN $14 THEN $15 ELSE is_active END
		WHERE id = $1
		RETURNING
			id,
			name,
			location,
			floor,
			capacity,
			equipment,
			description,
			is_active,
			created_at,
			updated_at;
	`

	var result Room

	err = r.pool.QueryRow(
		ctx,
		query,
		input.ID,
		input.Name.Set,
		input.Name.Value,
		input.Location.Set,
		input.Location.Value,
		input.Floor.Set,
		input.Floor.Value,
		input.Capacity.Set,
		input.Capacity.Value,
		input.Equipment.Set,
		equipmentJSON,
		input.Description.Set,
		input.Description.Value,
		input.IsActive.Set,
		input.IsActive.Value,
	).Scan(
		&result.ID,
		&result.Name,
		&result.Location,
		&result.Floor,
		&result.Capacity,
		&result.Equipment,
		&result.Description,
		&result.IsActive,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Room{}, ErrNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == roomsUniqueNamePerLocationConstraint {
			return Room{}, ErrAlreadyExists
		}

		return Room{}, fmt.Errorf("update room: %w", err)
	}

	return result, nil
}
