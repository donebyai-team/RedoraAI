BEGIN;

CREATE TABLE sub_reddits_leads
(
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    project_id uuid NOT NULL,
    author varchar(255) NOT NULL,
    keyword_id uuid NOT NULL, -- Table ID
    subreddit_id uuid NOT NULL, -- Table ID
    post_id varchar(255) NOT NULL,
    type character varying(255) NOT NULL, -- COMMENT, POST
    relevancy_score FLOAT NOT NULL,
    post_created_at timestamp NOT NULL, -- when the post was created on Reddit in UTC
    status varchar(255) NOT NULL DEFAULT 'NEW', -- NEW, NOT_RELEVANT, COMPLETED
    description TEXT NOT NULL,
    title TEXT, -- Comment won't have title
    metadata jsonb DEFAULT '{}'::jsonb, -- Store any metadata eg. no of comments, likes etc
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp
);

ALTER TABLE sub_reddits_leads ADD CONSTRAINT fk1_sub_reddits_leads FOREIGN KEY (project_id) REFERENCES projects (id);
ALTER TABLE sub_reddits_leads ADD CONSTRAINT fk2_sub_reddits_leads FOREIGN KEY (subreddit_id) REFERENCES sub_reddits (id);
ALTER TABLE sub_reddits_leads ADD CONSTRAINT fk3_sub_reddits_leads FOREIGN KEY (keyword_id) REFERENCES keywords (id);
CREATE UNIQUE INDEX idx1_sub_reddits_leads ON sub_reddits_leads (project_id, post_id);
CREATE INDEX idx2_sub_reddits_leads ON sub_reddits_leads (project_id);


CREATE TRIGGER trigger_record_changed_on_sub_reddits_leads
    BEFORE UPDATE
    ON sub_reddits_leads
    FOR EACH ROW
    EXECUTE PROCEDURE record_changed();

COMMIT;
