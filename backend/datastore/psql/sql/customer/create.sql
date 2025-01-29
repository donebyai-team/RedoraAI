INSERT INTO customers (name, phone, organization_id)
VALUES (:name, :phone, :organization_id) RETURNING id;