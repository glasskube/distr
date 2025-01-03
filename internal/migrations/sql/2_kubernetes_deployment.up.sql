ALTER TABLE DeploymentTarget
  ADD COLUMN namespace TEXT;

ALTER TABLE Deployment
  ADD COLUMN release_name TEXT,
  ADD COLUMN values_yaml BYTEA,
  ADD CONSTRAINT release_name_unique UNIQUE (deployment_target_id, release_name);
