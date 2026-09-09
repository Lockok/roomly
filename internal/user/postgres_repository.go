package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const usersEmailKeyConstraint = "users_email_key"

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func scanUser(row pgx.Row) (User, error) {
	var result User

	err := row.Scan(
		&result.ID,
		&result.Email,
		&result.PasswordHash,
		&result.FullName,
		&result.Role,
		&result.IsActive,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	return result, err
}

func (r *PostgresRepository) Create(ctx context.Context, input CreateInput) (User, error) {
	const query = `
		INSERT INTO users (email, password_hash, full_name, role)
		VALUES ($1, $2, $3, $4)
		RETURNING
			id,
			email,
			password_hash,
			full_name,
			role,
			is_active,
			created_at,
			updated_at;
	`

	result, err := scanUser(r.pool.QueryRow(
		ctx,
		query,
		input.Email,
		input.PasswordHash,
		input.FullName,
		input.Role,
	))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == usersEmailKeyConstraint {
			return User{}, ErrAlreadyExists
		}

		return User{}, fmt.Errorf("create user: %w", err)
	}

	return result, nil
}

func (r *PostgresRepository) List(ctx context.Context, filter ListFilter) ([]User, error) {
	const query = `
		SELECT
			id,
			email,
			password_hash,
			full_name,
			role,
			is_active,
			created_at,
			updated_at
		FROM users
		WHERE ($1::boolean IS NULL OR is_active = $1)
		ORDER BY full_name ASC, email ASC;
	`

	rows, err := r.pool.Query(ctx, query, filter.IsActive)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]User, 0)

	for rows.Next() {
		result, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}

		users = append(users, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}

	return users, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	const query = `
		SELECT
			id,
			email,
			password_hash,
			full_name,
			role,
			is_active,
			created_at,
			updated_at
		FROM users
		WHERE id = $1;
	`

	result, err := scanUser(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}

		return User{}, fmt.Errorf("get user by ID: %w", err)
	}

	return result, nil
}

func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (User, error) {
	const query = `
		SELECT
			id,
			email,
			password_hash,
			full_name,
			role,
			is_active,
			created_at,
			updated_at
		FROM users
		WHERE email = $1;
	`

	result, err := scanUser(r.pool.QueryRow(ctx, query, email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}

		return User{}, fmt.Errorf("get user by email: %w", err)
	}

	return result, nil
}
