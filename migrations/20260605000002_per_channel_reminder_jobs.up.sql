ALTER TABLE reminder_jobs
ADD COLUMN IF NOT EXISTS channel VARCHAR(20);

ALTER TABLE notifications
ADD COLUMN IF NOT EXISTS dedupe_key TEXT;

ALTER TABLE reminder_jobs
DROP CONSTRAINT IF EXISTS reminder_jobs_important_day_id_occurrence_date_offset_days_key;

WITH first_channels AS (
    SELECT
        id,
        COALESCE(
            (
                SELECT value
                FROM jsonb_array_elements_text(channels) WITH ORDINALITY AS channel_values(value, ordinality)
                WHERE value IN ('email', 'in_app', 'push')
                ORDER BY ordinality
                LIMIT 1
            ),
            'in_app'
        ) AS channel
    FROM reminder_jobs
)
UPDATE reminder_jobs AS jobs
SET channel = first_channels.channel
FROM first_channels
WHERE jobs.id = first_channels.id
  AND jobs.channel IS NULL;

WITH expanded AS (
    SELECT
        jobs.*,
        channel_values.value AS expanded_channel,
        ROW_NUMBER() OVER (
            PARTITION BY jobs.id, channel_values.value
            ORDER BY channel_values.ordinality
        ) AS channel_rank
    FROM reminder_jobs AS jobs
    CROSS JOIN LATERAL jsonb_array_elements_text(jobs.channels) WITH ORDINALITY AS channel_values(value, ordinality)
    WHERE channel_values.value IN ('email', 'in_app', 'push')
      AND channel_values.value <> jobs.channel
),
extra_jobs AS (
    SELECT *
    FROM expanded
    WHERE channel_rank = 1
)
INSERT INTO reminder_jobs (
    id,
    user_id,
    important_day_id,
    reminder_rule_id,
    occurrence_date,
    offset_days,
    channel,
    scheduled_at,
    status,
    attempts,
    last_error,
    locked_until,
    sent_at,
    created_at,
    updated_at
)
SELECT
    (
        SUBSTR(MD5(id::text || ':' || expanded_channel), 1, 8) || '-' ||
        SUBSTR(MD5(id::text || ':' || expanded_channel), 9, 4) || '-4' ||
        SUBSTR(MD5(id::text || ':' || expanded_channel), 14, 3) || '-a' ||
        SUBSTR(MD5(id::text || ':' || expanded_channel), 18, 3) || '-' ||
        SUBSTR(MD5(id::text || ':' || expanded_channel), 21, 12)
    )::uuid,
    user_id,
    important_day_id,
    reminder_rule_id,
    occurrence_date,
    offset_days,
    expanded_channel,
    scheduled_at,
    status,
    attempts,
    last_error,
    locked_until,
    sent_at,
    created_at,
    updated_at
FROM extra_jobs;

ALTER TABLE reminder_jobs
ALTER COLUMN channel SET NOT NULL;

ALTER TABLE reminder_jobs
ADD CONSTRAINT reminder_jobs_day_occurrence_offset_channel_key
UNIQUE (important_day_id, occurrence_date, offset_days, channel);

ALTER TABLE reminder_jobs
DROP COLUMN IF EXISTS channels;

CREATE UNIQUE INDEX IF NOT EXISTS idx_notifications_user_dedupe_key
ON notifications(user_id, dedupe_key)
WHERE dedupe_key IS NOT NULL;
