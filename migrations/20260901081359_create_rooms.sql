-- +goose Up
CREATE TABLE rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    location TEXT NOT NULL,
    floor INTEGER,
    capacity INTEGER NOT NULL,
    equipment JSONB NOT NULL DEFAULT '[]'::jsonb,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT rooms_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT rooms_location_not_empty CHECK (length(trim(location)) > 0),
    CONSTRAINT rooms_capacity_positive CHECK (capacity > 0),
    CONSTRAINT rooms_equipment_is_array CHECK (jsonb_typeof(equipment) = 'array'),
    CONSTRAINT rooms_unique_name_per_location UNIQUE (location, name)
);

-- +goose Down
DROP TABLE rooms;