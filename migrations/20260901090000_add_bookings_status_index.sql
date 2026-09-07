-- +goose Up
CREATE INDEX bookings_status_starts_at_idx
    ON bookings (status, starts_at);

-- +goose Down
DROP INDEX IF EXISTS bookings_status_starts_at_idx;
