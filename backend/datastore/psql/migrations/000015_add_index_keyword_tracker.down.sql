BEGIN;
DROP INDEX IF EXISTS idx_untracked_or_old_keyword_trackers;
DROP INDEX IF EXISTS idx_deleted_at_keyword_trackers;
COMMIT;