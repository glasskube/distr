ALTER TYPE DEPLOYMENT_STATUS_TYPE RENAME TO DEPLOYMENT_STATUS_TYPE_OLD;

CREATE TYPE DEPLOYMENT_STATUS_TYPE AS ENUM (
  'ok',
  'error',
  'progressing'
);

ALTER TABLE DeploymentRevisionStatus
  ALTER COLUMN type TYPE DEPLOYMENT_STATUS_TYPE USING CASE
    WHEN type = 'healthy' THEN 'ok'
    WHEN type = 'running' THEN 'ok'
    WHEN type = 'error' THEN 'error'
    WHEN type = 'progressing' THEN 'progressing'
  END::DEPLOYMENT_STATUS_TYPE;

DROP TYPE DEPLOYMENT_STATUS_TYPE_OLD;
