DROP INDEX IF EXISTS idx_user_sessions_user_active;

ALTER TABLE user_sessions
    DROP COLUMN IF EXISTS last_used_user_agent,
    DROP COLUMN IF EXISTS last_used_ip,
    DROP COLUMN IF EXISTS last_used_at,
    DROP COLUMN IF EXISTS created_user_agent,
    DROP COLUMN IF EXISTS created_ip,
    DROP COLUMN IF EXISTS revoked_reason;
