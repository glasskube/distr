CREATE TABLE ApplicationLicense (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  name TEXT NOT NULL,
  expires_at TIMESTAMP,
  application_id UUID NOT NULL REFERENCES Application (id) ON DELETE CASCADE,
  organization_id UUID NOT NULL REFERENCES Organization (id) ON DELETE CASCADE,
  owner_useraccount_id UUID REFERENCES UserAccount (id) ON DELETE CASCADE,
  registry_url TEXT,
  registry_username TEXT,
  registry_password TEXT,
  UNIQUE (name, organization_id),
  CHECK (
    (registry_url IS NULL AND registry_username IS NULL AND registry_password IS NULL)
    OR (registry_url IS NOT NULL AND registry_username IS NOT NULL AND registry_password IS NOT NULL)
  )
);

CREATE TABLE ApplicationLicense_ApplicationVersion (
  application_license_id UUID NOT NULL REFERENCES ApplicationLicense (id),
  application_version_id UUID NOT NULL REFERENCES ApplicationVersion (id),
  PRIMARY KEY (application_license_id, application_version_id)
);

ALTER TABLE Deployment
  ADD COLUMN application_license_id UUID REFERENCES ApplicationLicense (id);

CREATE INDEX fk_ApplicationLicense_application_id ON ApplicationLicense (application_id);
CREATE INDEX fk_ApplicationLicense_organization_id ON ApplicationLicense (organization_id);
CREATE INDEX fk_ApplicationLicense_owner_useraccount_id ON ApplicationLicense (owner_useraccount_id);
CREATE INDEX fk_Deployment_application_license_id ON Deployment (application_license_id);
CREATE INDEX fk_ApplicationLicense_ApplicationVersion_application_license_id
  ON ApplicationLicense_ApplicationVersion (application_license_id);
