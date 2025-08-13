SELECT *
FROM integrations
WHERE organization_id = :organization_id ORDER BY updated_at DESC;