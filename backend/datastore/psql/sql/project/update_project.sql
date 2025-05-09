UPDATE projects
SET
    name = :name,
    description = :description,
    customer_persona = :customer_persona,
    website = :website
WHERE
    id = :id AND organization_id = :organization_id;
