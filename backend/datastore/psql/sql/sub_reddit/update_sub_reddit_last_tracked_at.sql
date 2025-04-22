UPDATE sub_reddits set last_tracked_at = :last_tracked_at WHERE id = :id RETURNING *;
