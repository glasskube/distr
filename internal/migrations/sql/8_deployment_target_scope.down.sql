ALTER TABLE DeploymentTarget
  DROP CONSTRAINT scope_required,
  DROP CONSTRAINT namespace_required,
  DROP COLUMN scope;

DROP TYPE DEPLOYMENT_TARGET_SCOPE;
