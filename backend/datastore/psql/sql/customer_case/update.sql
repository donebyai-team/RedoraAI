UPDATE customer_sessions
SET status          = :status,
    last_call_status = :last_call_status,
    case_reason = :case_reason,
    next_scheduled_at = :next_scheduled_at,
    summary = :summary
WHERE id = :id;