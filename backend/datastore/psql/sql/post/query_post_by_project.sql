SELECT
    p.*,
    s.name AS "source.name",
    s.source_type AS "source.source_type",
    s.id AS "source.id"
FROM posts p
         JOIN sources s ON p.source_id = s.id
WHERE p.project_id = :project_id
ORDER BY p.created_at DESC;