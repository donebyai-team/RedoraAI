UPDATE sources set deleted_at = CURRENT_TIMESTAMP WHERE id = :id RETURNING *;
