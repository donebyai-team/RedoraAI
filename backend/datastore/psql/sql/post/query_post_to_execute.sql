SELECT *
FROM posts
WHERE status = :status
  AND schedule_at <= CURRENT_TIMESTAMP
ORDER BY created_at DESC;
