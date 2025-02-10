BEGIN;

ALTER TABLE conversations ADD COLUMN IF NOT EXISTS ai_decision jsonb DEFAULT '{}'::jsonb;

COMMIT;