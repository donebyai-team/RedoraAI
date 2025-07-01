SELECT *
FROM post_insights
WHERE project_id = :project_id
  AND relevancy_score > 0
  AND relevancy_score >= :relevancy_score
  AND (CAST(:start_datetime AS timestamp) IS NULL OR created_at >= :start_datetime)
  AND (CAST(:end_datetime AS timestamp) IS NULL OR created_at < :end_datetime)
ORDER BY l.post_created_at DESC, l.relevancy_score DESC
    LIMIT :limit
OFFSET :offset;