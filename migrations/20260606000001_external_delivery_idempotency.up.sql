ALTER TABLE device_tokens
ADD COLUMN IF NOT EXISTS provider VARCHAR(40) NOT NULL DEFAULT 'onesignal';

UPDATE device_tokens
SET provider = 'expo',
    active = false,
    updated_at = now()
WHERE token LIKE 'ExpoPushToken[%]'
   OR token LIKE 'ExponentPushToken[%]';

CREATE INDEX IF NOT EXISTS idx_device_tokens_user_active_provider
ON device_tokens(user_id, active, provider);

CREATE TABLE IF NOT EXISTS reminder_deliveries (
    id UUID PRIMARY KEY,
    reminder_job_id UUID NOT NULL REFERENCES reminder_jobs(id) ON DELETE CASCADE,
    channel VARCHAR(20) NOT NULL,
    target_id TEXT NOT NULL,
    provider VARCHAR(40) NOT NULL,
    idempotency_key TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    provider_message_id TEXT NOT NULL DEFAULT '',
    last_error TEXT NOT NULL DEFAULT '',
    sent_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE (reminder_job_id, channel, target_id),
    UNIQUE (idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_reminder_deliveries_job_channel
ON reminder_deliveries(reminder_job_id, channel);
