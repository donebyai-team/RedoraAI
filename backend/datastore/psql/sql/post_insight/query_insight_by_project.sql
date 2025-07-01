SELECT
    l.*,
    k.keyword AS "keyword.keyword",
    k.id AS "keyword.id",
    k.project_id AS "keyword.project_id",
    s.name AS "source.name",
    s.source_type AS "source.source_type",
    s.id AS "source.id"
FROM post_insights l
         JOIN keywords k ON l.keyword_id = k.id
         JOIN sources s ON l.source_id = s.id
WHERE l.project_id = :project_id
  AND l.relevancy_score > 0
  AND l.relevancy_score >= :relevancy_score
  AND (CAST(:start_datetime AS timestamp) IS NULL OR l.created_at >= :start_datetime)
  AND (CAST(:end_datetime AS timestamp) IS NULL OR l.created_at < :end_datetime)
ORDER BY l.created_at DESC, l.relevancy_score DESC
    LIMIT :limit
OFFSET :offset;