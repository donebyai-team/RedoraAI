SELECT *
FROM keyword_trackers
WHERE deleted_at IS NULL AND project_id = :project_id;