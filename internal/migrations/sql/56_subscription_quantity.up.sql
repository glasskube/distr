-- Update all NULL values to -1
UPDATE organization
SET subscription_user_account_quantity = -1
WHERE subscription_user_account_quantity IS NULL;

UPDATE organization
SET subscription_customer_organization_quantity = -1
WHERE subscription_customer_organization_quantity IS NULL;

-- Add NOT NULL constraints and set default values
ALTER TABLE organization
  ALTER COLUMN subscription_user_account_quantity SET DEFAULT -1,
  ALTER COLUMN subscription_user_account_quantity SET NOT NULL,
  ALTER COLUMN subscription_customer_organization_quantity SET DEFAULT -1,
  ALTER COLUMN subscription_customer_organization_quantity SET NOT NULL;
