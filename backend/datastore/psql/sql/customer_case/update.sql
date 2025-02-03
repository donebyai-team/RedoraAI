UPDATE customer_sessions
SET status          = :status,
    last_call_status = :last_call_status,
    next_scheduled_at = :next_scheduled_at,
    summary = :summary
WHERE id = :id;