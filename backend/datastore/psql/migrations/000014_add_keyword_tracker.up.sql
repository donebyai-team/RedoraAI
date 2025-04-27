BEGIN;

CREATE TABLE keyword_trackers
(
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    keyword_id uuid NOT NULL, -- Table ID
    source_id uuid NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb,
    last_tracked_at timestamp,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp,
    deleted_at timestamp
);

ALTER TABLE keyword_trackers ADD CONSTRAINT fk1_keyword_trackers FOREIGN KEY (keyword_id) REFERENCES keywords (id);
ALTER TABLE keyword_trackers ADD CONSTRAINT fk2_keyword_trackers FOREIGN KEY (source_id) REFERENCES sources (id);
CREATE UNIQUE INDEX idx1_keyword_trackers ON keyword_trackers (keyword_id, source_id);

CREATE TRIGGER trigger_record_changed_on_keyword_trackers
    BEFORE UPDATE
    ON keyword_trackers
    FOR EACH ROW
    EXECUTE PROCEDURE record_changed();

COMMIT;
