BEGIN;
-- Add schedule_at column with NULL default (or you can assign a default if needed)
ALTER TABLE lead_interactions
    ADD COLUMN schedule_at TIMESTAMP;

-- Create a multicolumn index to speed up queries with status + schedule_at + ordering
CREATE INDEX idx_lead_interactions_status_scheduleat
    ON lead_interactions (status, schedule_at);

COMMIT;
