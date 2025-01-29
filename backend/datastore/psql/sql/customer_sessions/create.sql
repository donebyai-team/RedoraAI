INSERT INTO customer_sessions (customer_id, organization_id, due_date, status)
VALUES (:customer_id, :organization_id, :due_date, :status) RETURNING id;