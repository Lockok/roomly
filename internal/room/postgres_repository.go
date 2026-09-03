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

func (r *PostgresRepository) List(ctx context.Context) ([]Room, error) {
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
		ORDER BY location ASC, name ASC;
	`
	rows, err := r.pool.Query(ctx, query)
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
