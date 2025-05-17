SELECT *
FROM lead_interactions
WHERE
    status = ANY(:statuses)
  AND schedule_at <= NOW()
    LIMIT 200;
