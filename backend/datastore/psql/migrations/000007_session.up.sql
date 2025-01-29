BEGIN;

CREATE TABLE customer_sessions
(
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    customer_id uuid NOT NULL,
    organization_id uuid NOT NULL,
    external_id varchar(255),
    due_date  timestamp NOT NULL,
    status character varying(255) NOT NULL DEFAULT 'QUEUED',
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp
);

ALTER TABLE customer_sessions ADD CONSTRAINT fk1_customer_sessions FOREIGN KEY (customer_id) REFERENCES customers (id);
ALTER TABLE customer_sessions ADD CONSTRAINT fk2_customer_sessions FOREIGN KEY (organization_id) REFERENCES organizations (id);

CREATE TRIGGER trigger_record_changed_on_customer_sessions
    BEFORE UPDATE
    ON customer_sessions
    FOR EACH ROW
    EXECUTE PROCEDURE record_changed();

COMMIT;
