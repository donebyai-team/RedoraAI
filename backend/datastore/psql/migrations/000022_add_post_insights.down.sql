BEGIN;

DROP TRIGGER IF EXISTS trigger_record_changed_on_post_insights ON post_insights;

DROP INDEX IF EXISTS idx_post_insights_proj_score_created;
DROP INDEX IF EXISTS idx_post_insights_post_project;

ALTER TABLE post_insights DROP CONSTRAINT IF EXISTS fk1_post_insights;
DROP TABLE IF EXISTS post_insights;

COMMIT;