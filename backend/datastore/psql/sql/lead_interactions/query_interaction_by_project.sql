SELECT *
FROM lead_interactions
WHERE project_id = :project_id
  AND status = :status
  AND created_at >= :start_date
  AND created_at < (CAST(:end_date AS date) + INTERVAL '1 day');;