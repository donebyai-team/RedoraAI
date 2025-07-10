SELECT *
FROM posts
WHERE status = :status
  AND schedule_at <= NOW()
ORDER BY created_at DESC;
