INSERT INTO conversations (customer_case_id, from_phone, provider)
VALUES (:customer_case_id, :from_phone, :provider) RETURNING id;