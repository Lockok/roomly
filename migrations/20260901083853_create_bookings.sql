-- +goose Up
CREATE TYPE booking_status AS ENUM ('confirmed', 'cancelled');

CREATE TABLE bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES rooms(id),
    organizer_id UUID NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    description TEXT,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    status booking_status NOT NULL DEFAULT 'confirmed',
    attendees_count INTEGER NOT NULL DEFAULT 1,
    cancelled_at TIMESTAMPTZ,
    cancelled_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT bookings_title_not_empty CHECK (length(trim(title)) > 0),
    CONSTRAINT bookings_time_range_valid CHECK (ends_at > starts_at),
    CONSTRAINT bookings_attendees_count_positive CHECK (attendees_count > 0),
    CONSTRAINT bookings_cancellation_data_valid CHECK (
        (status = 'confirmed' AND cancelled_at IS NULL AND cancelled_by IS NULL)
        OR
        (status = 'cancelled' AND cancelled_at IS NOT NULL AND cancelled_by IS NOT NULL)
    )
);

CREATE INDEX bookings_room_id_starts_at_idx
    ON bookings (room_id, starts_at);

CREATE INDEX bookings_organizer_id_starts_at_idx
    ON bookings (organizer_id, starts_at);

CREATE TRIGGER bookings_set_updated_at
BEFORE UPDATE ON bookings
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

ALTER TABLE bookings
ADD CONSTRAINT bookings_no_overlapping_confirmed
EXCLUDE USING gist (
    room_id WITH =,
    tstzrange(starts_at, ends_at, '[)') WITH &&
)
WHERE (status = 'confirmed');

-- +goose Down
ALTER TABLE bookings
    DROP CONSTRAINT IF EXISTS bookings_no_overlapping_confirmed;

DROP TRIGGER IF EXISTS bookings_set_updated_at ON bookings;

DROP INDEX IF EXISTS bookings_organizer_id_starts_at_idx;
DROP INDEX IF EXISTS bookings_room_id_starts_at_idx;

DROP TABLE bookings;
DROP TYPE booking_status;