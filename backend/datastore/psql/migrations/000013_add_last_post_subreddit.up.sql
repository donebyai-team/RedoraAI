BEGIN;

ALTER TABLE sub_reddits ADD COLUMN IF NOT EXISTS last_post_created_at timestamp;
ALTER TABLE sub_reddits ADD COLUMN IF NOT EXISTS metadata jsonb DEFAULT '{}'::jsonb;

COMMIT;