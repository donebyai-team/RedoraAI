INSERT INTO keywords (keyword, organization_id)
VALUES (:keyword, :organization_id) RETURNING id;
