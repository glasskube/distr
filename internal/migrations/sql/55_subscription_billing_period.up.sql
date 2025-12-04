CREATE TYPE SubscriptionPeriod AS ENUM ('monthly', 'yearly');

ALTER TABLE Organization
  ADD COLUMN subscription_period SubscriptionPeriod NOT NULL DEFAULT 'monthly';
