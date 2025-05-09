SELECT kt.*
FROM keyword_trackers kt
    JOIN projects p ON p.id = kt.project_id
    JOIN keywords k ON k.id = kt.keyword_id
    JOIN sources s ON s.id = kt.source_id
WHERE kt.deleted_at IS NULL
  AND p.is_active = true
  AND k.deleted_at IS NULL
  AND s.deleted_at IS NULL
  AND (kt.last_tracked_at IS NULL OR kt.last_tracked_at < NOW() - INTERVAL '24 hours');
