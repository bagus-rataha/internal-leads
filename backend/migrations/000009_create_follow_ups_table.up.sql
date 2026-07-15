CREATE TABLE IF NOT EXISTS follow_ups (
    id            UUID PRIMARY KEY,
    lead_id       UUID NOT NULL REFERENCES leads (id),
    note          TEXT NOT NULL,
    created_by_id UUID NOT NULL REFERENCES users (id),
    created_at    TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_follow_ups_lead_id_created_at ON follow_ups (lead_id, created_at);
