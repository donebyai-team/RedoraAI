SELECT *
FROM sub_reddits_leads
WHERE project_id = :project_id
  AND status = :status
ORDER BY created_at DESC;