DELETE FROM sub_reddits WHERE id = :id RETURNING *;
