BEGIN;

ALTER TABLE integrations
    RENAME COLUMN config TO plain_text_config;

ALTER TABLE integrations
    ADD COLUMN encrypted_config VARCHAR NOT NULL DEFAULT '';

COMMIT;
