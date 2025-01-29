UPDATE customer_sessions
SET status          = :status,
    external_id     = :external_id
WHERE id = :id;