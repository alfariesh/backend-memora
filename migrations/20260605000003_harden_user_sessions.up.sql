ALTER TABLE user_sessions
    ADD COLUMN IF NOT EXISTS revoked_reason TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS created_ip TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS created_user_agent TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS last_used_ip TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS last_used_user_agent TEXT NOT NULL DEFAULT '';

UPDATE user_sessions
SET revoked_at = now(),
    revoked_reason = 'migration',
    updated_at = now()
WHERE revoked_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_user_sessions_user_active
    ON user_sessions(user_id, created_at DESC)
    WHERE revoked_at IS NULL;
