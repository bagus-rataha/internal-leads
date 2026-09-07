ALTER TABLE leads DROP CONSTRAINT chk_leads_status;
ALTER TABLE leads ADD CONSTRAINT chk_leads_status CHECK (status IN ('BARU','FOLLOW_UP','HANDOFF_ODOO','LOST'));
DROP INDEX idx_leads_survey_at;
ALTER TABLE leads DROP COLUMN survey_at;
