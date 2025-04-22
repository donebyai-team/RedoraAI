UPDATE sub_reddits set deleted_at = CURRENT_TIMESTAMP WHERE id = :id RETURNING *;
