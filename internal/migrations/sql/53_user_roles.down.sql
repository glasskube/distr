ALTER TABLE Organization_UserAccount DROP COLUMN user_role;

DROP TYPE USER_ROLE;

CREATE TYPE USER_ROLE AS ENUM ('vendor', 'customer');

ALTER TABLE Organization_UserAccount
  ADD COLUMN user_role USER_ROLE;

UPDATE Organization_UserAccount
  SET user_role = CASE WHEN (customer_organization_id IS NOT NULL)
    THEN 'customer'::USER_ROLE
    ELSE 'vendor'::USER_ROLE
  END;

ALTER TABLE Organization_UserAccount
  ALTER COLUMN user_role SET NOT NULL;

-- Recreate the constraint, since it is dropped implicitly by recreating the user_role column in the "up" migration
ALTER TABLE Organization_UserAccount
  ADD CONSTRAINT customer_organization_id_null_check
    CHECK ((user_role = 'customer') = (customer_organization_id IS NOT NULL));
