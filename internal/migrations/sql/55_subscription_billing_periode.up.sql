CREATE TYPE SubscriptionPeriode AS ENUM ('monthly', 'yearly');

ALTER TABLE Organization
  ADD COLUMN subscription_periode SubscriptionPeriode NOT NULL DEFAULT 'monthly';
