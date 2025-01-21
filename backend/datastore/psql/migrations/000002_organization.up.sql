BEGIN;

CREATE TABLE organizations
(
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    name varchar(255) NOT NULL,
    feature_flags jsonb DEFAULT '{}'::jsonb,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp
);

CREATE UNIQUE INDEX IF NOT EXISTS idx1_organizations on organizations (name);

CREATE TRIGGER trigger_record_changed_on_organizations
    BEFORE UPDATE
    ON organizations
    FOR EACH ROW
    EXECUTE PROCEDURE record_changed();

COMMIT;
