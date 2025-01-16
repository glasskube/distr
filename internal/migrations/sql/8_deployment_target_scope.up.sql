CREATE TYPE DEPLOYMENT_TARGET_SCOPE AS ENUM ('cluster', 'namespace');

ALTER TABLE DeploymentTarget
  ADD COLUMN scope DEPLOYMENT_TARGET_SCOPE;

UPDATE DeploymentTarget SET scope = 'namespace' WHERE type = 'kubernetes';

ALTER TABLE DeploymentTarget
  ADD CONSTRAINT scope_required CHECK ((type = 'docker') = (scope IS NULL)),
  ADD CONSTRAINT namespace_required CHECK ((type = 'docker') = (namespace IS NULL));
