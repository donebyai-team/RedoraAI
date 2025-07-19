BEGIN;

-- Drop the new unique index that includes reference_id
DROP INDEX IF EXISTS idx1_integrations_new;

-- Restore the original unique index on (organization_id, type)
CREATE UNIQUE INDEX idx1_integrations
    ON integrations (organization_id, type);

-- Drop the reference_id column
ALTER TABLE integrations
DROP COLUMN IF EXISTS reference_id;

COMMIT;
