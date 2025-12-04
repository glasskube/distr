-- Remove NOT NULL constraints and default values
ALTER TABLE organization
  ALTER COLUMN subscription_user_account_quantity DROP NOT NULL,
  ALTER COLUMN subscription_user_account_quantity DROP DEFAULT,
  ALTER COLUMN subscription_customer_organization_quantity DROP NOT NULL,
  ALTER COLUMN subscription_customer_organization_quantity DROP DEFAULT;

-- Set -1 values back to NULL
UPDATE organization
SET subscription_user_account_quantity = NULL
WHERE subscription_user_account_quantity = -1;

UPDATE organization
SET subscription_customer_organization_quantity = NULL
WHERE subscription_customer_organization_quantity = -1;
