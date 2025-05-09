SELECT *
FROM projects
WHERE LOWER(name) = LOWER(:name) AND organization_id = :organization_id;