SELECT *
FROM integrations
WHERE organization_id = :organization_id AND type = :type;