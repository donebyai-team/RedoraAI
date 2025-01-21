BEGIN;
-- enable module to generate universally unique identifiers
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- capture record change
CREATE OR REPLACE FUNCTION record_changed() RETURNS TRIGGER
    LANGUAGE plpgsql
AS
$$
BEGIN
    NEW.updated_at := (CURRENT_TIMESTAMP at time zone 'utc');
    RETURN NEW;
END;
$$;

COMMIT;
