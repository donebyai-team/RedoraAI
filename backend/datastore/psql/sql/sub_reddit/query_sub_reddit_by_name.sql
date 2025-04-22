SELECT *
FROM sub_reddits
WHERE name = :name AND project_id = :project_id AND deleted_at IS NULL;
