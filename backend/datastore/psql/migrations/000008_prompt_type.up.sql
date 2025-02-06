BEGIN;

CREATE TABLE prompt_types (
    name character varying(255) NOT NULL,
    description TEXT,
    organization_id uuid NOT NULL,
    config jsonb DEFAULT '{}'::jsonb  NOT NULL,
    created_at             timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at             timestamp,
    PRIMARY KEY (name)
);


ALTER TABLE prompt_types ADD CONSTRAINT fk1_prompt_types FOREIGN KEY (organization_id) REFERENCES organizations (id);
CREATE UNIQUE INDEX idx1_prompt_types ON prompt_types (name);

CREATE TRIGGER trigger_record_changed_on_prompt_types
    BEFORE UPDATE
    ON prompt_types
    FOR EACH ROW
    EXECUTE PROCEDURE record_changed();

COMMIT;