UPDATE posts
SET
    status = :status,
    schedule_at = :schedule_at,
    updated_at = CURRENT_TIMESTAMP
WHERE id = :id;
