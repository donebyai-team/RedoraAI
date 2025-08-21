INSERT INTO integrations (
    id,
    organization_id,
    type,
    reference_id,
    encrypted_config,
    plain_text_config,
    state,
    metadata
)
VALUES (
           :id,
           :organization_id,
           :type,
           :reference_id,  -- this should be passed explicitly
           :encrypted_config,
           :plain_text_config,
           :state,
           :metadata
       )
    ON CONFLICT (organization_id, type, reference_id)
DO UPDATE SET
    plain_text_config = excluded.plain_text_config,
           encrypted_config = excluded.encrypted_config,
           metadata = excluded.metadata,
           state = excluded.state
           RETURNING *;
