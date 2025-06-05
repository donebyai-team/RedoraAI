SELECT *
FROM lead_interactions
WHERE
    status = :status AND
    from_user = :from_user AND
    to_user = :to_user AND
    type = :type AND
    project_id = :project_id
