CREATE TYPE DEPLOYMENT_TYPE AS ENUM ('docker', 'kubernetes');

CREATE TYPE USER_ROLE AS ENUM ('vendor', 'customer');

CREATE TABLE IF NOT EXISTS Organization (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS UserAccount (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  email TEXT NOT NULL UNIQUE,
  password_hash BYTEA,
  password_salt BYTEA,
  name TEXT
);

CREATE TABLE IF NOT EXISTS Organization_UserAccount (
  organization_id UUID NOT NULL REFERENCES Organization (id) ON DELETE CASCADE,
  user_account_id UUID NOT NULL REFERENCES UserAccount (id) ON DELETE CASCADE,
  user_role USER_ROLE NOT NULL,
  PRIMARY KEY (organization_id, user_account_id)
);

CREATE INDEX IF NOT EXISTS fk_Organization_UserAccount_organization_id ON Organization_UserAccount (organization_id);
CREATE INDEX IF NOT EXISTS fk_Organization_UserAccount_user_account_id ON Organization_UserAccount (user_account_id);

CREATE TABLE IF NOT EXISTS Application (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  name TEXT NOT NULL,
  type DEPLOYMENT_TYPE NOT NULL,
  organization_id UUID NOT NULL REFERENCES Organization (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS fk_Application_organization_id ON Application (organization_id);

CREATE TABLE IF NOT EXISTS ApplicationVersion (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  name TEXT NOT NULL,
  compose_file_data BYTEA,
  application_id UUID NOT NULL REFERENCES Application (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS fk_ApplicationVersion_application_id ON ApplicationVersion (application_id);

CREATE TABLE IF NOT EXISTS DeploymentTarget (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  created_by_user_account_id UUID NOT NULL REFERENCES UserAccount (id),
  name TEXT NOT NULL,
  type DEPLOYMENT_TYPE NOT NULL,
  geolocation_lat FLOAT,
  geolocation_lon FLOAT,
  access_key_salt BYTEA,
  access_key_hash BYTEA,
  organization_id UUID NOT NULL REFERENCES Organization (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS fk_DeploymentTarget_organization_id ON DeploymentTarget (organization_id);

CREATE TABLE IF NOT EXISTS DeploymentTargetStatus (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  deployment_target_id UUID NOT NULL REFERENCES DeploymentTarget (id) ON DELETE CASCADE,
  message TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS fk_DeploymentTargetStatus_deployment_target_id ON DeploymentTargetStatus (deployment_target_id);

CREATE TABLE IF NOT EXISTS Deployment (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  deployment_target_id UUID NOT NULL REFERENCES DeploymentTarget (id),
  application_version_id UUID NOT NULL REFERENCES ApplicationVersion (id)
);

CREATE INDEX IF NOT EXISTS fk_Deployment_deployment_target_id ON Deployment (deployment_target_id);
CREATE INDEX IF NOT EXISTS fk_Deployment_application_version_id ON Deployment (application_version_id);

CREATE TABLE IF NOT EXISTS DeploymentStatus (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  deployment_id UUID NOT NULL REFERENCES Deployment (id) ON DELETE CASCADE,
  message TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS fk_DeploymentStatus_deployment_target_id ON DeploymentStatus (deployment_id);
