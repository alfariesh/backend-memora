DROP INDEX IF EXISTS idx_notifications_user_dedupe_key;

ALTER TABLE notifications
DROP COLUMN IF EXISTS dedupe_key;

ALTER TABLE reminder_jobs
ADD COLUMN IF NOT EXISTS channels JSONB NOT NULL DEFAULT '[]'::jsonb;

UPDATE reminder_jobs
SET channels = jsonb_build_array(channel)
WHERE channel IS NOT NULL;

ALTER TABLE reminder_jobs
DROP CONSTRAINT IF EXISTS reminder_jobs_day_occurrence_offset_channel_key;

DELETE FROM reminder_jobs AS duplicate
USING reminder_jobs AS kept
WHERE duplicate.important_day_id = kept.important_day_id
  AND duplicate.occurrence_date = kept.occurrence_date
  AND duplicate.offset_days = kept.offset_days
  AND duplicate.id::text > kept.id::text;

ALTER TABLE reminder_jobs
DROP COLUMN IF EXISTS channel;

ALTER TABLE reminder_jobs
ADD CONSTRAINT reminder_jobs_important_day_id_occurrence_date_offset_days_key
UNIQUE (important_day_id, occurrence_date, offset_days);
