INSERT INTO customer_sessions (customer_id, organization_id, due_date, prompt_type)
VALUES (:customer_id, :organization_id, :due_date, :prompt_type) RETURNING id;