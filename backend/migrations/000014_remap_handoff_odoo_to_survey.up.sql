UPDATE leads
SET survey_at = COALESCE(survey_at, updated_at),
    status = 'SURVEY'
WHERE status = 'HANDOFF_ODOO';

ALTER TABLE leads DROP CONSTRAINT chk_leads_status;
ALTER TABLE leads ADD CONSTRAINT chk_leads_status CHECK (status IN (
  'BARU','FOLLOW_UP','SURVEY','SALES_CONFIRMATION','REGISTRASI',
  'INSTALASI','TRIAL','INVOICE_BULANAN','LOST'
));
