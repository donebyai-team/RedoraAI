BEGIN;
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS fk1_subscriptions;
DROP TABLE IF EXISTS subscriptions;
COMMIT;
