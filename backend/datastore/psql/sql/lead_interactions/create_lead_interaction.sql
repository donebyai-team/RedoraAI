INSERT INTO lead_interactions (
    project_id,
    lead_id,
    type,
    from_user,
    to_user,
    reason,
    metadata,
    schedule_at)
VALUES (
           :project_id,
           :lead_id,
           :type,
           :from_user,
           :to_user,
           :reason,
           :metadata,
           :schedule_at)
    RETURNING id;