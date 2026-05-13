CREATE TABLE IF NOT EXISTS important_days (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    type VARCHAR(40) NOT NULL,
    person_name VARCHAR(255) NOT NULL DEFAULT '',
    relationship VARCHAR(100) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    event_year INTEGER,
    event_month SMALLINT NOT NULL,
    event_day SMALLINT NOT NULL,
    recurrence VARCHAR(20) NOT NULL DEFAULT 'yearly',
    timezone VARCHAR(64) NOT NULL DEFAULT 'Asia/Jakarta',
    reminder_time CHAR(5) NOT NULL DEFAULT '09:00',
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_important_days_user_id ON important_days(user_id);
CREATE INDEX IF NOT EXISTS idx_important_days_user_type ON important_days(user_id, type);

CREATE TABLE IF NOT EXISTS reminder_rules (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    important_day_id UUID NOT NULL REFERENCES important_days(id) ON DELETE CASCADE,
    offset_days INTEGER NOT NULL,
    channels JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_reminder_rules_day ON reminder_rules(important_day_id);

CREATE TABLE IF NOT EXISTS reminder_jobs (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    important_day_id UUID NOT NULL REFERENCES important_days(id) ON DELETE CASCADE,
    reminder_rule_id UUID REFERENCES reminder_rules(id) ON DELETE SET NULL,
    occurrence_date DATE NOT NULL,
    offset_days INTEGER NOT NULL,
    channels JSONB NOT NULL DEFAULT '[]'::jsonb,
    scheduled_at TIMESTAMP NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT NOT NULL DEFAULT '',
    locked_until TIMESTAMP,
    sent_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE (important_day_id, occurrence_date, offset_days)
);

CREATE INDEX IF NOT EXISTS idx_reminder_jobs_due ON reminder_jobs(status, scheduled_at);
CREATE INDEX IF NOT EXISTS idx_reminder_jobs_user ON reminder_jobs(user_id);

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    important_day_id UUID REFERENCES important_days(id) ON DELETE SET NULL,
    type VARCHAR(40) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    data JSONB NOT NULL DEFAULT '{}'::jsonb,
    read_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_created ON notifications(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications(user_id, read_at) WHERE read_at IS NULL;

CREATE TABLE IF NOT EXISTS device_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    platform VARCHAR(40) NOT NULL,
    name VARCHAR(255) NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE (user_id, token)
);

CREATE INDEX IF NOT EXISTS idx_device_tokens_user_active ON device_tokens(user_id, active);
