SELECT *
FROM lead_interactions
WHERE project_id = :project_id
  AND status = :status
  AND (CAST(:start_datetime AS timestamp) IS NULL OR created_at >= :start_datetime)
  AND (CAST(:end_datetime AS timestamp) IS NULL OR created_at < :end_datetime)