BEGIN;

CREATE TABLE conversations
(
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    customer_id uuid NOT NULL,
    from_phone varchar(255) NOT NULL,
    organization_id uuid NOT NULL,
    status varchar(255) NOT NULL,
    summary TEXT DEFAULT '',
    call_duration integer DEFAULT 0,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp
);

ALTER TABLE conversations ADD CONSTRAINT fk1_conversations FOREIGN KEY (organization_id) REFERENCES organizations (id);
ALTER TABLE conversations ADD CONSTRAINT fk2_conversations FOREIGN KEY (customer_id) REFERENCES customers (id);


CREATE TRIGGER trigger_record_changed_on_conversations
    BEFORE UPDATE
    ON conversations
    FOR EACH ROW
    EXECUTE PROCEDURE record_changed();

COMMIT;
