CREATE SEQUENCE IF NOT EXISTS lead_code_seq;

CREATE TABLE IF NOT EXISTS leads (
    id                UUID PRIMARY KEY,
    code              TEXT UNIQUE NOT NULL,
    status            TEXT NOT NULL DEFAULT 'BARU',
    lost_reason       TEXT,
    owner_id          UUID NOT NULL REFERENCES users (id),
    created_by_id     UUID NOT NULL REFERENCES users (id),
    company_name      TEXT NOT NULL,
    business_field    TEXT,
    website           TEXT,
    province_id       INTEGER REFERENCES provinces (id),
    city_id           INTEGER REFERENCES cities (id),
    district_id       INTEGER REFERENCES districts (id),
    village_id        INTEGER REFERENCES villages (id),
    zip_id            INTEGER REFERENCES zips (id),
    rt                TEXT,
    rw                TEXT,
    street            TEXT,
    pic_name          TEXT,
    pic_position      TEXT,
    office_phone      TEXT,
    mobile_phone      TEXT,
    email             TEXT,
    service_type_id   UUID REFERENCES service_types (id),
    capacity_mbps     INTEGER,
    existing_isp      TEXT,
    price             NUMERIC,
    other_services    TEXT,
    lead_source_id    UUID REFERENCES lead_sources (id),
    last_follow_up_at TIMESTAMPTZ,
    created_at        TIMESTAMPTZ,
    updated_at        TIMESTAMPTZ,
    deleted_at        TIMESTAMPTZ,
    CONSTRAINT chk_leads_status CHECK (status IN ('BARU','FOLLOW_UP','HANDOFF_ODOO','LOST')),
    CONSTRAINT chk_leads_lost_reason CHECK (status <> 'LOST' OR lost_reason IS NOT NULL)
);
CREATE INDEX IF NOT EXISTS idx_leads_owner_id ON leads (owner_id);
CREATE INDEX IF NOT EXISTS idx_leads_status ON leads (status);
CREATE INDEX IF NOT EXISTS idx_leads_last_follow_up_at ON leads (last_follow_up_at);
CREATE INDEX IF NOT EXISTS idx_leads_created_at ON leads (created_at);
CREATE INDEX IF NOT EXISTS idx_leads_city_id ON leads (city_id);
