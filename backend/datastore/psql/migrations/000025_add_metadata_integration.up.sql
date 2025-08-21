BEGIN;
ALTER TABLE integrations ADD COLUMN metadata jsonb DEFAULT '{}'::jsonb;
COMMIT;