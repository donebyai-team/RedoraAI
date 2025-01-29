BEGIN;

CREATE TABLE conversations
(
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    session_id uuid NOT NULL,
    from_phone varchar(255) NOT NULL,
    status character varying(255) NOT NULL DEFAULT 'QUEUED',
    summary TEXT DEFAULT '',
    voice_provider character varying(255) NOT NULL,
    call_duration integer DEFAULT 0,
    recording_url TEXT,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp
);

ALTER TABLE conversations ADD CONSTRAINT fk1_conversations FOREIGN KEY (session_id) REFERENCES customer_sessions (id);

CREATE TRIGGER trigger_record_changed_on_conversations
    BEFORE UPDATE
    ON conversations
    FOR EACH ROW
    EXECUTE PROCEDURE record_changed();

COMMIT;
