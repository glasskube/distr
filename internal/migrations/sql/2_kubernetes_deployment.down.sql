ALTER TABLE DeploymentTarget
  DROP COLUMN namespace;

ALTER TABLE Deployment
  DROP CONSTRAINT release_name_unique,
  DROP COLUMN release_name,
  DROP COLUMN values_yaml;
