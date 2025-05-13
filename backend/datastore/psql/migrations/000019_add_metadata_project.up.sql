BEGIN;
ALTER TABLE projects ADD COLUMN metadata jsonb DEFAULT '{}'::jsonb;
COMMIT;