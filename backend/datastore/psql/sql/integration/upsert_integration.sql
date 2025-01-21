INSERT INTO integrations (
    id,
    organization_id,
    type,
    encrypted_config,
    plain_text_config,
    state
)
VALUES (
    :id,
    :organization_id,
    :type,
    :encrypted_config,
    :plain_text_config,
    :state
)
ON CONFLICT (organization_id, type)
DO UPDATE SET
    plain_text_config = excluded.plain_text_config,
    encrypted_config = excluded.encrypted_config,
    state = excluded.state
RETURNING *;