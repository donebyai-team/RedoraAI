INSERT INTO conversations (customer_session_id, from_phone, status, due_date, provider)
VALUES (:customer_session_id, :from_phone, :status, :due_date, :provider) RETURNING id;