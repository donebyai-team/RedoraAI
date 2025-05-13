BEGIN;

CREATE TABLE keywords
(
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    keyword varchar(255) NOT NULL,
    project_id uuid NOT NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp,
    deleted_at timestamp
);

ALTER TABLE keywords ADD CONSTRAINT fk_keywords FOREIGN KEY (project_id) REFERENCES projects (id);
CREATE INDEX idx_keywords_project_id_deleted_at_null ON keywords (project_id) WHERE deleted_at IS NULL;


CREATE TRIGGER trigger_record_changed_on_keywords
    BEFORE UPDATE
    ON keywords
    FOR EACH ROW
    EXECUTE PROCEDURE record_changed();

COMMIT;
