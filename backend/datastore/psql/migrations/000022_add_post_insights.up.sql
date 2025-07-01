BEGIN;

CREATE TABLE post_insights
(
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    project_id uuid NOT NULL,
    post_id varchar(255) NOT NULL,
    keyword_id uuid NOT NULL, -- Table ID
    source_id uuid NOT NULL, -- Table ID
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
ALTER TABLE leads ADD CONSTRAINT fk2_post_insights FOREIGN KEY (source_id) REFERENCES sources (id);
ALTER TABLE leads ADD CONSTRAINT fk3_post_insights FOREIGN KEY (keyword_id) REFERENCES keywords (id);

CREATE INDEX idx_post_insights_proj_score_created
    ON post_insights (project_id, relevancy_score, created_at);

CREATE INDEX idx_post_insights_post_project
    ON post_insights (post_id, project_id);

CREATE TRIGGER trigger_record_changed_on_post_insights
    BEFORE UPDATE
    ON post_insights
    FOR EACH ROW
    EXECUTE PROCEDURE record_changed();

COMMIT;
