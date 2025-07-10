UPDATE posts
SET
    status = :status,
    post_id = :post_id,
    reason = :reason,
    updated_at = NOW()
WHERE id = :id
  AND status = 'SCHEDULED';
