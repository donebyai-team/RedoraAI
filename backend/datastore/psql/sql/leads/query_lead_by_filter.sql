SELECT *
FROM leads
WHERE project_id = :project_id
  AND relevancy_score >= :relevancy_score
  AND (:source_ids = '{}' OR source_id = ANY(:source_ids))
  AND status = :status
ORDER BY post_created_at DESC;