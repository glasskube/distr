-- Recreate owner_useraccount_id columns in ApplicationLicense and ArtifactLicense
ALTER TABLE ApplicationLicense
  ADD COLUMN owner_useraccount_id UUID REFERENCES UserAccount(id);

ALTER TABLE ArtifactLicense
  ADD COLUMN owner_useraccount_id UUID REFERENCES UserAccount(id);

-- Populate owner_useraccount_id from customer_organization_id
-- Set owner_useraccount_id to the first user (alphabetically by join date) in that customer organization
UPDATE ApplicationLicense AS al
SET owner_useraccount_id = (
  SELECT ou.user_account_id
  FROM Organization_UserAccount ou
  WHERE ou.customer_organization_id = al.customer_organization_id
  ORDER BY ou.created_at ASC
  LIMIT 1
);

UPDATE ArtifactLicense AS al
SET owner_useraccount_id = (
  SELECT ou.user_account_id
  FROM Organization_UserAccount ou
  WHERE ou.customer_organization_id = al.customer_organization_id
  ORDER BY ou.created_at ASC
  LIMIT 1
);

-- Remove the constraint first
ALTER TABLE Organization_UserAccount
  DROP CONSTRAINT customer_organization_id_null_check;

-- Remove the foreign key columns (data will be lost)
ALTER TABLE ArtifactLicense
  DROP COLUMN customer_organization_id;

ALTER TABLE ApplicationLicense
  DROP COLUMN customer_organization_id;

ALTER TABLE DeploymentTarget
  DROP COLUMN customer_organization_id;

ALTER TABLE Organization_UserAccount
  DROP COLUMN customer_organization_id;

-- Drop the CustomerOrganization table
DROP TABLE CustomerOrganization;
