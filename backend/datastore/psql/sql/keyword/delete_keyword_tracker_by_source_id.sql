UPDATE keyword_trackers set deleted_at = CURRENT_TIMESTAMP WHERE source_id = :source_id RETURNING *;
