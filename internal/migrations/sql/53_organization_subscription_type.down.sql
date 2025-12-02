ALTER TABLE Organization
  DROP COLUMN subscription_user_account_quantity,
  DROP COLUMN subscription_customer_organization_quantity,
  DROP COLUMN stripe_subscription_id,
  DROP COLUMN stripe_customer_id,
  DROP COLUMN subscription_ends_at,
  DROP COLUMN subscription_type;

DROP TYPE SubscriptionType;
