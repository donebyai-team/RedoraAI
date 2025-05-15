BEGIN;

ALTER TABLE projects DROP COLUMN IF EXISTS is_active;
DROP INDEX IF EXISTS idx_projects_id_active;
DROP INDEX IF EXISTS idx_keywords_id_active;
DROP INDEX IF EXISTS idx_sources_id_active;

COMMIT;
