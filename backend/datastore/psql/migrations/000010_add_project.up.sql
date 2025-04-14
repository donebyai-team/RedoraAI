BEGIN;

CREATE TABLE projects
(
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    name varchar(255) NOT NULL,
    organization_id uuid NOT NULL,
    description TEXT NOT NULL,
    customer_persona TEXT NOT NULL,
    website TEXT NOT NULL,
    goals TEXT NOT NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp
);

ALTER TABLE projects ADD CONSTRAINT fk_projects FOREIGN KEY (organization_id) REFERENCES organizations (id);


CREATE TRIGGER trigger_record_changed_on_projects
    BEFORE UPDATE
    ON projects
    FOR EACH ROW
    EXECUTE PROCEDURE record_changed();

COMMIT;
