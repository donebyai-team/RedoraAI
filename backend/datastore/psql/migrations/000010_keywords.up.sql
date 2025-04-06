BEGIN;

CREATE TABLE keywords
(
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    keyword varchar(255) NOT NULL,
    organization_id uuid NOT NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp
);

ALTER TABLE keywords ADD CONSTRAINT fk_keywords FOREIGN KEY (organization_id) REFERENCES organizations (id);


CREATE TRIGGER trigger_record_changed_on_keywords
    BEFORE UPDATE
    ON keywords
    FOR EACH ROW
    EXECUTE PROCEDURE record_changed();

COMMIT;
