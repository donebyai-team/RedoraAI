SELECT *
FROM sub_reddits
WHERE last_tracked_at < NOW() - INTERVAL '24 hours';
