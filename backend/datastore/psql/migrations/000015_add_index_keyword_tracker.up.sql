CREATE INDEX idx_untracked_or_old_keyword_trackers
    ON keyword_trackers(last_tracked_at)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_deleted_at_keyword_trackers ON keyword_trackers(deleted_at);
