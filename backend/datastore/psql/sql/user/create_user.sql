INSERT INTO users (auth0_id, email, email_verified, organization_id, role, state)
VALUES (:auth0_id, :email, :email_verified, :organization_id, :role, :state) RETURNING id;