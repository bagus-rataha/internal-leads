CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS lead_sources (
    id         UUID PRIMARY KEY,
    name       TEXT UNIQUE NOT NULL,
    is_active  BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_lead_sources_deleted_at ON lead_sources (deleted_at);

INSERT INTO lead_sources (id, name, created_at, updated_at) VALUES
    (gen_random_uuid(), 'Database Eksisting', now(), now()),
    (gen_random_uuid(), 'Google', now(), now()),
    (gen_random_uuid(), 'AI', now(), now()),
    (gen_random_uuid(), 'Sistem Internal', now(), now()),
    (gen_random_uuid(), 'Referensi Pelanggan', now(), now()),
    (gen_random_uuid(), 'Komunitas Business', now(), now()),
    (gen_random_uuid(), 'Event/Pameran', now(), now()),
    (gen_random_uuid(), 'Media Sosial', now(), now()),
    (gen_random_uuid(), 'LinkedIn Sales Navigator', now(), now()),
    (gen_random_uuid(), 'Upselling/Cross Selling', now(), now());
