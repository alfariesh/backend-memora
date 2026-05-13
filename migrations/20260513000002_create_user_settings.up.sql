CREATE TABLE IF NOT EXISTS user_settings (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    timezone VARCHAR(64) NOT NULL DEFAULT 'Asia/Jakarta',
    reminder_time CHAR(5) NOT NULL DEFAULT '09:00',
    notification_channels JSONB NOT NULL DEFAULT '["email","in_app","push"]'::jsonb,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);
