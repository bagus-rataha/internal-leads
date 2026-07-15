CREATE TABLE IF NOT EXISTS sales_teams (
    id         UUID PRIMARY KEY,
    name       TEXT NOT NULL,
    is_active  BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_sales_teams_deleted_at ON sales_teams (deleted_at);
