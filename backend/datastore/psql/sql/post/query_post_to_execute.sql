SELECT *
FROM posts
WHERE status = :status
  AND schedule_at <= CURRENT_TIMESTAMP AND deleted_at IS NULL;
