SELECT *
FROM customer_cases
WHERE status = ANY(:status)
  AND (last_call_status is NULL OR last_call_status = ANY(:last_call_status))
  AND (next_scheduled_at IS NULL OR next_scheduled_at <= :current_time);