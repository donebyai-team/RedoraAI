SELECT *
FROM posts
WHERE id = :id AND deleted_at IS NULL;