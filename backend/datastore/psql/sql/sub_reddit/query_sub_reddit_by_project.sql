SELECT *
FROM sub_reddits
WHERE project_id = :project_id AND deleted_at IS NULL;