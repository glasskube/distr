-- Remove 'community' from SubscriptionType enum
-- Note: PostgreSQL does not support removing values from an enum type directly.
-- We need to recreate the type without the 'community' value.

UPDATE Organization
  SET subscription_type = 'trial'
  WHERE subscription_type = 'community';

ALTER TABLE Organization
  ALTER COLUMN subscription_type DROP DEFAULT;

CREATE TYPE SubscriptionType_new AS ENUM ('starter', 'pro', 'enterprise', 'trial');

ALTER TABLE Organization
  ALTER COLUMN subscription_type TYPE SubscriptionType_new
  USING subscription_type::text::SubscriptionType_new;

DROP TYPE SubscriptionType;
ALTER TYPE SubscriptionType_new RENAME TO SubscriptionType;

ALTER TABLE Organization
  ALTER COLUMN subscription_type SET DEFAULT 'trial';
