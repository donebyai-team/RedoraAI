BEGIN;

CREATE TABLE post_insights
(
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    project_id uuid NOT NULL,
    post_id varchar(255) NOT NULL,
    source_type varchar(255) NOT NULL,
    relevancy_score FLOAT NOT NULL,
    topic text NOT NULL,
    sentiment text NOT NULL,
    highlights text NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp
);

-- Assuming projects and posts are tracked in other tables
ALTER TABLE post_insights ADD CONSTRAINT fk1_post_insights FOREIGN KEY (project_id) REFERENCES projects (id);

-- Indexes to support efficient querying
CREATE INDEX idx_post_insights_project_created_at ON post_insights (project_id, created_at);
CREATE INDEX idx_post_insights_project_id ON post_insights (project_id);


-- Optional: audit trigger
CREATE TRIGGER trigger_record_changed_on_post_insights
    BEFORE UPDATE
    ON post_insights
    FOR EACH ROW
    EXECUTE PROCEDURE record_changed();

COMMIT;
