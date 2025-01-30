INSERT INTO customers (first_name, last_name, phone, organization_id)
VALUES (:first_name, :last_name, :organization_id) RETURNING id;