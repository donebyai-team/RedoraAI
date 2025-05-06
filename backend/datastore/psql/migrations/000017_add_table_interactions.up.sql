BEGIN;

CREATE TABLE lead_interactions
(
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    project_id uuid NOT NULL,
    lead_id uuid NOT NULL,
    type character varying(255) NOT NULL, -- DM, COMMENT, LIKE
    from_user varchar(255) NOT NULL,
    to_user varchar(255) NOT NULL,
    status varchar(255) NOT NULL DEFAULT 'CREATED', -- CREATED, SENT, FAILED
    reason TEXT,
    metadata jsonb DEFAULT '{}'::jsonb, -- Store any metadata
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp
);

ALTER TABLE lead_interactions ADD CONSTRAINT fk1_lead_interactions FOREIGN KEY (project_id) REFERENCES projects (id);
ALTER TABLE lead_interactions ADD CONSTRAINT fk2_lead_interactions FOREIGN KEY (lead_id) REFERENCES leads (id);

CREATE TRIGGER trigger_record_changed_on_lead_interactions
    BEFORE UPDATE
    ON lead_interactions
    FOR EACH ROW
    EXECUTE PROCEDURE record_changed();

COMMIT;