ALTER TABLE Organization
  DROP COLUMN subscription_user_account_quantity,
  DROP COLUMN subscription_customer_organization_quantity,
  DROP COLUMN subscription_external_id,
  DROP COLUMN subscription_ends_at,
  DROP COLUMN subscription_type;

DROP TYPE SubscriptionType;
