-- Re-widens the CHECK to allow HANDOFF_ODOO again. The status/survey_at remap
-- is NOT reversed: which rows were HANDOFF_ODOO is not recoverable. Rows that
-- were remapped stay status='SURVEY' with survey_at set. (Same lossy-down
-- pattern noted for enum-expansion migrations in the design doc section 8.)
ALTER TABLE leads DROP CONSTRAINT chk_leads_status;
ALTER TABLE leads ADD CONSTRAINT chk_leads_status CHECK (status IN (
  'BARU','FOLLOW_UP','SURVEY','SALES_CONFIRMATION','REGISTRASI',
  'INSTALASI','TRIAL','INVOICE_BULANAN','LOST','HANDOFF_ODOO'
));
