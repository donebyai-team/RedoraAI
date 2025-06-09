UPDATE projects
SET
    is_active = :is_active
WHERE
    organization_id = :organization_id;
