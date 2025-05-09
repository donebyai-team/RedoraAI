BEGIN;

CREATE TABLE sources
(
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    external_id varchar(255),
    name varchar(255) NOT NULL,
    source_type varchar(255) NOT NULL,
    description TEXT NOT NULL,
    project_id uuid NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp,
    deleted_at timestamp
);

ALTER TABLE sources ADD CONSTRAINT fk_sources FOREIGN KEY (project_id) REFERENCES projects (id);
CREATE UNIQUE INDEX idx1_sources ON sources (project_id, external_id, deleted_at);
CREATE INDEX idx2_sources ON sources (project_id, external_id);


CREATE TRIGGER trigger_record_changed_on_sources
    BEFORE UPDATE
    ON sources
    FOR EACH ROW
    EXECUTE PROCEDURE record_changed();

COMMIT;
