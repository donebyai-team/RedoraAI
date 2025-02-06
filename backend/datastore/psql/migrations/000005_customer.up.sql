BEGIN;

CREATE TABLE customers
(
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    first_name varchar(255) NOT NULL,
    last_name varchar(255) NOT NULL,
    phone varchar(255) NOT NULL,
    organization_id uuid NOT NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp
);

ALTER TABLE customers ADD CONSTRAINT fk1_customers FOREIGN KEY (organization_id) REFERENCES organizations (id);
CREATE UNIQUE INDEX idx1_customers ON customers (organization_id, phone);
CREATE INDEX idx2_customers ON customers (organization_id, phone);

CREATE TRIGGER trigger_record_changed_on_customers
    BEFORE UPDATE
    ON customers
    FOR EACH ROW
    EXECUTE PROCEDURE record_changed();

COMMIT;
