BEGIN;
ALTER TABLE integrations
    ADD COLUMN reference_id VARCHAR;

UPDATE integrations
SET reference_id = CASE
                       WHEN type = 'REDDIT' THEN plain_text_config->>'name'
                       WHEN type = 'REDDIT_DM_LOGIN' THEN plain_text_config->>'username'
                       ELSE NULL
    END;

DROP INDEX IF EXISTS idx1_integrations;

CREATE UNIQUE INDEX idx1_integrations_new
    ON integrations (organization_id, type, reference_id);

COMMIT;
