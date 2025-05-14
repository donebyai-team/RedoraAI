BEGIN;

CREATE TABLE subscriptions (
     id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
     organization_id uuid NOT NULL,
     plan_id text NOT NULL,
     status text NOT NULL,
     amount NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
     metadata jsonb DEFAULT '{}'::jsonb,
     external_id text,
     created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
     updated_at timestamp,
     expires_at timestamp NOT NULL
);

ALTER TABLE subscriptions
    ADD CONSTRAINT fk1_subscriptions FOREIGN KEY (organization_id) REFERENCES organizations (id);

CREATE TRIGGER trigger_record_changed_on_subscriptions
    BEFORE UPDATE ON subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION record_changed();

COMMIT;
