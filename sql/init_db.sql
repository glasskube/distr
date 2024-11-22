CREATE TYPE DEPLOYMENT_TYPE AS ENUM ('docker', 'kubernetes');

CREATE TABLE IF NOT EXISTS Organization (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS Application (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  name TEXT NOT NULL,
  type DEPLOYMENT_TYPE NOT NULL
);

CREATE TABLE IF NOT EXISTS Organization_Application (
  organization_id UUID NOT NULL REFERENCES Organization(id),
  application_id UUID NOT NULL REFERENCES Application(id)
);

CREATE TABLE IF NOT EXISTS ApplicationVersion (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  name TEXT NOT NULL,
  compose_file_data BYTEA,
  application_id UUID NOT NULL REFERENCES Application(id)
);

CREATE TABLE IF NOT EXISTS DeploymentTarget (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  name TEXT NOT NULL,
  type DEPLOYMENT_TYPE NOT NULL
);

CREATE TABLE IF NOT EXISTS Organization_DeploymentTarget (
  organization_id UUID NOT NULL REFERENCES Organization(id),
  user_account_id UUID NOT NULL REFERENCES DeploymentTarget(id)
);

CREATE TABLE IF NOT EXISTS DeploymentTargetStatus (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  deployment_target_id UUID NOT NULL REFERENCES DeploymentTarget(id),
  message TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS Deployment (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  deployment_target_id UUID NOT NULL REFERENCES DeploymentTarget(id),
  application_version_id UUID NOT NULL REFERENCES ApplicationVersion(id)
);

CREATE TABLE IF NOT EXISTS DeploymentStatus (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  deployment_id UUID NOT NULL REFERENCES Deployment(id),
  message TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS UserAccount (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  email TEXT NOT NULL UNIQUE,
  password_hash BYTEA NOT NULL,
  name TEXT
);

CREATE TABLE IF NOT EXISTS Organization_UserAccount (
  organization_id UUID NOT NULL REFERENCES Organization(id),
  user_account_id UUID NOT NULL REFERENCES UserAccount(id)
);
