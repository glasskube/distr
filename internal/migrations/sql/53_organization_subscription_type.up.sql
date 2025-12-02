CREATE TYPE SubscriptionType AS ENUM ('starter', 'pro', 'enterprise', 'trial');

ALTER TABLE Organization
  ADD COLUMN subscription_type SubscriptionType NOT NULL DEFAULT 'trial',
  ADD COLUMN subscription_ends_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() + interval '1 month'),
  ADD COLUMN stripe_subscription_id TEXT,
  ADD COLUMN stripe_customer_id TEXT,
  ADD COLUMN subscription_customer_organization_quantity INTEGER,
  ADD COLUMN subscription_user_account_quantity INTEGER,
  ALTER COLUMN features SET DEFAULT '{licensing}';

UPDATE Organization SET features = '{licensing}';
