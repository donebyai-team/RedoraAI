BEGIN;

ALTER TABLE sub_reddits DROP COLUMN IF EXISTS last_post_created_at;
ALTER TABLE sub_reddits DROP COLUMN IF EXISTS metadata;
COMMIT;