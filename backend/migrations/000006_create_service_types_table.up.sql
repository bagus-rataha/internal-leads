CREATE TABLE IF NOT EXISTS service_types (
    id         UUID PRIMARY KEY,
    name       TEXT UNIQUE NOT NULL,
    is_active  BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_service_types_deleted_at ON service_types (deleted_at);

INSERT INTO service_types (id, name, created_at, updated_at) VALUES
    (gen_random_uuid(), 'Dedicated', now(), now()),
    (gen_random_uuid(), 'Broadband', now(), now());
