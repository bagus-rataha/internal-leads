CREATE TABLE IF NOT EXISTS users (
    id         UUID PRIMARY KEY,
    email      TEXT NOT NULL,
    password   TEXT NOT NULL,
    name       TEXT NOT NULL,
    role       TEXT DEFAULT 'user',
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);
