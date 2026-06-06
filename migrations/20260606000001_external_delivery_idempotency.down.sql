DROP INDEX IF EXISTS idx_reminder_deliveries_job_channel;
DROP TABLE IF EXISTS reminder_deliveries;

DROP INDEX IF EXISTS idx_device_tokens_user_active_provider;

ALTER TABLE device_tokens
DROP COLUMN IF EXISTS provider;
