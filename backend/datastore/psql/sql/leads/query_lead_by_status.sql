SELECT *
FROM leads
WHERE project_id = :project_id
  AND status = :status
ORDER BY created_at DESC;