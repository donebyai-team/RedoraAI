UPDATE lead_interactions
SET status = 'PROCESSING', updated_at = CURRENT_TIMESTAMP
WHERE id = :id AND status = 'CREATED';
