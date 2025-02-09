INSERT INTO customer_cases (customer_id, organization_id, due_date, prompt_type)
VALUES (:customer_id, :organization_id, :due_date, :prompt_type) RETURNING id;