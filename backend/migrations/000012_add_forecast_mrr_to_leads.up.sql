ALTER TABLE leads ADD COLUMN forecast_mrr NUMERIC(14,2);
ALTER TABLE leads ADD CONSTRAINT leads_forecast_mrr_nonneg CHECK (forecast_mrr IS NULL OR forecast_mrr >= 0);
