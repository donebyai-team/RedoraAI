BEGIN;

CREATE TABLE integrations
(
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    organization_id uuid  NOT NULL,
    type character varying(255) NOT NULL,
    state character varying(255) NOT NULL,
    plain_text_config jsonb DEFAULT '{}'::jsonb  NOT NULL,
    encrypted_config VARCHAR NOT NULL DEFAULT ''
    created_at             timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at             timestamp
);

CREATE UNIQUE INDEX idx1_integrations ON integrations (organization_id,type);
ALTER TABLE integrations ADD CONSTRAINT fk1_integrations FOREIGN KEY (organization_id) REFERENCES organizations (id);

CREATE TRIGGER trigger_record_changed_on_integrations
    BEFORE UPDATE
    ON integrations
    FOR EACH ROW
    EXECUTE PROCEDURE record_changed();


COMMIT;
