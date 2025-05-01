SELECT COUNT(*)
FROM leads
WHERE project_id = :project_id
  AND relevancy_score >= :relevancy_score
  AND created_at >= :start_date
  AND created_at < (CAST(:end_date AS date) + INTERVAL '1 day');