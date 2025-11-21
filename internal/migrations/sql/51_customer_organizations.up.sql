CREATE TABLE CustomerOrganization (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  organization_id UUID NOT NULL REFERENCES Organization(id),
  image_id UUID REFERENCES File(id),
  name TEXT NOT NULL
);

ALTER TABLE Organization_UserAccount
  ADD COLUMN customer_organization_id UUID REFERENCES CustomerOrganization(id);

ALTER TABLE DeploymentTarget
  ADD COLUMN customer_organization_id UUID REFERENCES CustomerOrganization(id);

ALTER TABLE ApplicationLicense
  ADD COLUMN customer_organization_id UUID REFERENCES CustomerOrganization(id);

ALTER TABLE ArtifactLicense
  ADD COLUMN customer_organization_id UUID REFERENCES CustomerOrganization(id);

CREATE INDEX fk_CustomerOrganization_organization_id ON CustomerOrganization(organization_id);
CREATE INDEX fk_Organization_UserAccount_customer_organization_id ON Organization_UserAccount(customer_organization_id);
CREATE INDEX fk_DeploymentTarget_customer_organization_id ON DeploymentTarget(customer_organization_id);
CREATE INDEX fk_ApplicationLicense_customer_organization_id ON ApplicationLicense(customer_organization_id);
CREATE INDEX fk_ArtifactLicense_customer_organization_id ON ArtifactLicense(customer_organization_id);

INSERT INTO CustomerOrganization (organization_id, name)
SELECT ou.organization_id, u.email
FROM Organization_UserAccount ou
JOIN UserAccount u ON ou.user_account_id = u.id
WHERE ou.user_role = 'customer';

UPDATE Organization_UserAccount AS ou
SET customer_organization_id = co.id
FROM UserAccount u
JOIN CustomerOrganization co ON co.name = u.email
WHERE ou.user_account_id = u.id
  AND ou.organization_id = co.organization_id;

UPDATE DeploymentTarget AS dt
SET customer_organization_id = ou.customer_organization_id
FROM Organization_UserAccount ou
WHERE dt.created_by_user_account_id = ou.user_account_id
  AND dt.organization_id = ou.organization_id
  AND ou.customer_organization_id IS NOT NULL;

UPDATE ApplicationLicense AS al
SET customer_organization_id = ou.customer_organization_id
FROM Organization_UserAccount ou
WHERE al.owner_useraccount_id = ou.user_account_id
  AND al.organization_id = ou.organization_id
  AND ou.customer_organization_id IS NOT NULL;

UPDATE ArtifactLicense AS al
SET customer_organization_id = ou.customer_organization_id
FROM Organization_UserAccount ou
WHERE al.owner_useraccount_id = ou.user_account_id
  AND al.organization_id = ou.organization_id
  AND ou.customer_organization_id IS NOT NULL;

ALTER TABLE Organization_UserAccount
  ADD CONSTRAINT customer_organization_id_null_check
    CHECK ((user_role = 'customer') = (customer_organization_id IS NOT NULL));

ALTER TABLE ApplicationLicense
  DROP COLUMN owner_useraccount_id;

ALTER TABLE ArtifactLicense
  DROP COLUMN owner_useraccount_id;
