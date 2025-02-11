SELECT *
FROM customer_cases
WHERE status = ANY(:status)
  AND (last_call_status IS NULL OR last_call_status = ANY(:last_call_status))
  AND (next_scheduled_at IS NULL OR next_scheduled_at <= :current_time)
  AND (
        next_scheduled_at IS NOT NULL
        OR updated_at IS NULL
        OR updated_at <= NOW() - INTERVAL '10 minutes'
    )
ORDER BY created_at DESC;