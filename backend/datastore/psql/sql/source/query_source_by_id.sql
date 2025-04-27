SELECT *
FROM sources
WHERE id = :id AND deleted_at IS NULL;