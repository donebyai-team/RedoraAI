UPDATE keyword_trackers set last_tracked_at = :last_tracked_at WHERE id = :id RETURNING *;
