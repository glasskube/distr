CREATE TABLE DeploymentRevision (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  deployment_id UUID NOT NULL REFERENCES Deployment (id) ON DELETE CASCADE,
  application_version_id UUID NOT NULL REFERENCES ApplicationVersion (id) ON DELETE CASCADE,
  values_yaml BYTEA
);

-- TODO here: migrate
-- for each deployment target:
--    for each deployment: create a deployment revision, and delete the deployment if its not the very first one of the target

-- alternative to migration logic: simply truncate the deployment table and don't migrate deployments

ALTER TABLE Deployment
  DROP COLUMN application_version_id,
  DROP COLUMN values_yaml;
-- release_name stays in deployment and the constraint can be added again
-- TODO re-check null handling in unique constraint
ALTER TABLE Deployment ADD CONSTRAINT release_name_unique UNIQUE (deployment_target_id, release_name);
