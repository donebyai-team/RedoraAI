BEGIN;

CREATE TABLE posts
(
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    project_id uuid NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    source_id uuid NOT NULL, -- Table ID
    status varchar(255) NOT NULL DEFAULT 'CREATED', -- CREATED, SENT, FAILED, SCHEDULED
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    reason TEXT,
    reference_id uuid, -- Insight table ID optional
    post_id varchar(255), -- Post ID in the source system (e.g., Reddit, Twitter)
    schedule_at timestamp,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp
);

-- Assuming projects and posts are tracked in other tables
ALTER TABLE posts ADD CONSTRAINT fk1_posts FOREIGN KEY (project_id) REFERENCES projects (id);
ALTER TABLE posts ADD CONSTRAINT fk2_posts FOREIGN KEY (source_id) REFERENCES sources (id);
ALTER TABLE posts ADD CONSTRAINT fk3_posts FOREIGN KEY (reference_id) REFERENCES post_insights (id);

CREATE INDEX idx_post_projects
    ON posts (project_id);

CREATE INDEX idx_posts_schedule_at_not_null ON posts(schedule_at)
    WHERE schedule_at IS NOT NULL;

CREATE TRIGGER trigger_record_changed_on_posts
    BEFORE UPDATE
    ON posts
    FOR EACH ROW
    EXECUTE PROCEDURE record_changed();

COMMIT;
