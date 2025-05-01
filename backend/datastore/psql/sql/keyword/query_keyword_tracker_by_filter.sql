SELECT *
FROM keyword_trackers
WHERE (deleted_at IS NULL AND (last_tracked_at IS NULL OR last_tracked_at < NOW() - INTERVAL '24 hours'));