INSERT INTO lead_interactions (
    project_id,
    lead_id,
    type,
    from_user,
    to_user,
    reason,
    metadata)
VALUES (
           :project_id,
           :lead_id,
           :type,
           :from_user,
           :to_user,
           :reason,
           :metadata)
    RETURNING id;