BEGIN;

DROP TABLE IF EXISTS keywords;
DROP INDEX IF EXISTS fk_keywords;
DROP INDEX IF EXISTS idx_keywords_project_id_deleted_at_null;

COMMIT;