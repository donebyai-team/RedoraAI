BEGIN;

DROP TRIGGER IF EXISTS trigger_record_changed_on_posts ON post_insights;

DROP INDEX IF EXISTS idx_post_projects;
DROP INDEX IF EXISTS idx_posts_schedule_at_not_null;

ALTER TABLE posts DROP CONSTRAINT IF EXISTS fk1_posts;
ALTER TABLE posts DROP CONSTRAINT IF EXISTS fk2_posts;
ALTER TABLE posts DROP CONSTRAINT IF EXISTS fk3_posts;

DROP TABLE IF EXISTS posts;

COMMIT;