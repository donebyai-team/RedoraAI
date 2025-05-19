SELECT COUNT(*)
FROM leads
WHERE project_id = :project_id
  AND relevancy_score >= :relevancy_score
  AND (CAST(:start_datetime AS timestamp) IS NULL OR created_at >= :start_datetime)
  AND (CAST(:end_datetime AS timestamp) IS NULL OR created_at < :end_datetime)