ALTER TABLE users ADD COLUMN team_id UUID NULL REFERENCES sales_teams (id);
ALTER TABLE users ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE users ALTER COLUMN role SET DEFAULT 'SALES';
ALTER TABLE users ADD CONSTRAINT chk_users_role CHECK (role IN ('SALES','LEADER','ADMIN_SALES','SU'));
ALTER TABLE users ADD CONSTRAINT chk_users_team_by_role CHECK (
    (role IN ('SALES','LEADER') AND team_id IS NOT NULL) OR
    (role IN ('ADMIN_SALES','SU') AND team_id IS NULL)
);
CREATE INDEX IF NOT EXISTS idx_users_team_id ON users (team_id);
