INSERT INTO subscriptions (
    organization_id,
    plan_id,
    status,
    metadata,
    external_id,
    amount,
    expires_at
)
VALUES (
           :organization_id,
           :plan_id,
           :status,
           :metadata,
           :external_id,
           :amount,
           :expires_at
       )
RETURNING id;
