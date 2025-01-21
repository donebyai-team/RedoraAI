BEGIN;

-- users
CREATE TABLE users
(
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    auth0_id character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    email_verified boolean NOT NULL,
    organization_id uuid,
    role character varying(255),
    state character varying(255),
    created_at             timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at             timestamp
);

CREATE UNIQUE INDEX idx1_users ON users USING btree (email);
CREATE INDEX idx2_users ON users USING btree (organization_id);
ALTER TABLE users ADD CONSTRAINT fk1_users FOREIGN KEY (organization_id) REFERENCES organizations (id);

CREATE TRIGGER trigger_record_changed_on_users
    BEFORE UPDATE
    ON users
    FOR EACH ROW
    EXECUTE PROCEDURE record_changed();

COMMIT;
