BEGIN;

CREATE TABLE customer_cases
(
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    customer_id uuid NOT NULL,
    organization_id uuid NOT NULL,
    due_date  timestamp NOT NULL,
    prompt_type character varying(255) NOT NULL,
    status character varying(255) NOT NULL DEFAULT 'CREATED',
    summary TEXT DEFAULT '',
    next_scheduled_at timestamp,
    last_call_status character varying(255),
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp
);

ALTER TABLE customer_cases ADD CONSTRAINT fk1_customer_cases FOREIGN KEY (customer_id) REFERENCES customers (id);
ALTER TABLE customer_cases ADD CONSTRAINT fk2_customer_cases FOREIGN KEY (organization_id) REFERENCES organizations (id);

CREATE TRIGGER trigger_record_changed_on_customer_cases
    BEFORE UPDATE
    ON customer_cases
    FOR EACH ROW
    EXECUTE PROCEDURE record_changed();

COMMIT;
