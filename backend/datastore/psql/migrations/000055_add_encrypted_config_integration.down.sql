BEGIN;

ALTER TABLE integrations
    DROP COLUMN encrypted_config;

ALTER TABLE integrations
    RENAME COLUMN plain_text_config TO config;

COMMIT;
