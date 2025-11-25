ALTER TABLE Organization
  RENAME COLUMN subscription_external_id TO stripe_subscription_id;
ALTER TABLE Organization
  ADD COLUMN stripe_customer_id TEXT;
