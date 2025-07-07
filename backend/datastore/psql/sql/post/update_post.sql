UPDATE posts
SET
    title = :title,
    description = :description,
    status = :status,
    metadata = :metadata,
    reason = :reason,
    reference_id = :reference_id,
    updated_at = CURRENT_TIMESTAMP
WHERE id = :id;
