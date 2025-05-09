BEGIN;
ALTER TABLE leads ADD COLUMN IF NOT EXISTS intents TEXT[];
CREATE INDEX idx1_intents_gin_leads ON leads USING GIN (intents);
COMMIT;