BEGIN;

CREATE TABLE sub_reddits
(
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    subreddit_id varchar(255) NOT NULL,
    url varchar(255) NOT NULL,
    name varchar(255) NOT NULL,
    description varchar(5000) NOT NULL,
    organization_id uuid NOT NULL,
    subreddit_created_at timestamp NOT NULL,
    last_tracked_at timestamp,
    subscribers integer,
    title varchar(255),
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp
);

ALTER TABLE sub_reddits ADD CONSTRAINT fk_sub_reddits FOREIGN KEY (organization_id) REFERENCES organizations (id);
CREATE UNIQUE INDEX idx1_subreddit ON sub_reddits (organization_id, subreddit_id);
CREATE INDEX idx2_subreddit ON sub_reddits (organization_id, subreddit_id);


CREATE TRIGGER trigger_record_changed_on_sub_reddits
    BEFORE UPDATE
    ON sub_reddits
    FOR EACH ROW
    EXECUTE PROCEDURE record_changed();

COMMIT;
