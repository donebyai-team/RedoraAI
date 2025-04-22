BEGIN;

CREATE TABLE subreddit_trackers
(
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    project_id uuid NOT NULL,
    subreddit_id uuid NOT NULL,
    keyword_id uuid NOT NULL,
    newest_tracked_post varchar(255),
    oldest_tracked_post varchar(255),
    last_tracked_at timestamp,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp
);

ALTER TABLE subreddit_trackers ADD CONSTRAINT fk_subreddit_trackers1 FOREIGN KEY (project_id) REFERENCES projects (id);
ALTER TABLE subreddit_trackers ADD CONSTRAINT fk_subreddit_trackers2 FOREIGN KEY (subreddit_id) REFERENCES sub_reddits (id);
ALTER TABLE subreddit_trackers ADD CONSTRAINT fk_subreddit_trackers3 FOREIGN KEY (keyword_id) REFERENCES keywords (id);
CREATE UNIQUE INDEX idx1_subreddit_trackers ON subreddit_trackers (project_id, subreddit_id, keyword_id);

CREATE TRIGGER trigger_record_changed_on_subreddit_trackers
    BEFORE UPDATE
    ON subreddit_trackers
    FOR EACH ROW
    EXECUTE PROCEDURE record_changed();

COMMIT;
