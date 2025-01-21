UPDATE users
SET email           = :email,
    email_verified  = :email_verified,
    organization_id = :organization_id,
    role            = :role,
    state           = :state
WHERE id = :id;