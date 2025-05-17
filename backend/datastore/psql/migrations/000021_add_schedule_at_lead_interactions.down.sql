BEGIN;
-- Drop the index
DROP INDEX IF EXISTS idx_lead_interactions_status_scheduleat;

-- Remove the column
ALTER TABLE lead_interactions
DROP COLUMN IF EXISTS schedule_at;
COMMIT;