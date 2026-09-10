package report

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) RoomUsage(ctx context.Context, input RoomUsageInput) ([]RoomUsage, error) {
	const query = `
	SELECT
		r.id,
		r.name,
		r.location,
		COUNT(b.id)::integer AS bookings_count,
		COALESCE(
			SUM(
				EXTRACT(
					EPOCH FROM (
						LEAST(b.ends_at, $2::timestamptz)
						- GREATEST(b.starts_at, $1::timestamptz)
					)
				) / 60
			) FILTER (WHERE b.id IS NOT NULL),
			0
		)::integer AS booked_minutes,
		COALESCE(
			ROUND(
				100.0
				* COALESCE(
					SUM(
						EXTRACT(
							EPOCH FROM (
								LEAST(b.ends_at, $2::timestamptz)
								- GREATEST(b.starts_at, $1::timestamptz)
							)
						) / 60
					) FILTER (WHERE b.id IS NOT NULL),
					0
				)
				/ (EXTRACT(EPOCH FROM ($2::timestamptz - $1::timestamptz)) / 60),
				2
			),
			0
		)::float8 AS utilization_rate
	FROM rooms AS r
	LEFT JOIN bookings AS b
		ON b.room_id = r.id
		AND b.status = 'confirmed'
		AND b.ends_at > $1::timestamptz
		AND b.starts_at < $2::timestamptz
	GROUP BY r.id, r.name, r.location
	ORDER BY utilization_rate DESC, r.location ASC, r.name ASC;
`

	rows, err := r.pool.Query(ctx, query, input.From, input.To)
	if err != nil {
		return nil, fmt.Errorf("query room usage report: %w", err)
	}
	defer rows.Close()

	result := make([]RoomUsage, 0)

	for rows.Next() {
		var item RoomUsage

		if err := rows.Scan(
			&item.RoomID,
			&item.RoomName,
			&item.Location,
			&item.BookingsCount,
			&item.BookedMinutes,
			&item.UtilizationRate,
		); err != nil {
			return nil, fmt.Errorf("scan room usage report: %w", err)
		}

		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate room usage report: %w", err)
	}

	return result, nil
}
