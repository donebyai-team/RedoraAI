SELECT *
FROM conversations
WHERE customer_case_id = :customer_case_id AND call_status != 'CREATED';