ALTER TABLE Organization
  RENAME COLUMN stripe_subscription_id TO subscription_external_id;
ALTER TABLE Organization
  DROP COLUMN stripe_customer_id;
