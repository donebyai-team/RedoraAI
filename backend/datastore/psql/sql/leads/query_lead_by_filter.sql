SELECT
    l.*,
    k.keyword AS "keyword.keyword",
    k.id AS "keyword.id",
    k.project_id AS "keyword.project_id"
FROM leads l
JOIN keywords k ON l.keyword_id = k.id
WHERE l.project_id = :project_id
  AND l.relevancy_score > 0
  AND l.relevancy_score >= :relevancy_score
  AND (:source_ids = '{}' OR l.source_id = ANY(:source_ids))
  AND l.status = :status
  AND (CAST(:start_datetime AS timestamp) IS NULL OR l.created_at >= :start_datetime)
  AND (CAST(:end_datetime AS timestamp) IS NULL OR l.created_at < :end_datetime)
ORDER BY l.post_created_at DESC, l.relevancy_score DESC
LIMIT :limit
OFFSET :offset;