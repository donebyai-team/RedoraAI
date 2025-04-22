SELECT *
FROM keywords
WHERE id = :id AND deleted_at IS NULL;