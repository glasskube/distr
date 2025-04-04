-- Credits: https://stackoverflow.com/a/25812436

ALTER TYPE DEPLOYMENT_STATUS_TYPE
  RENAME TO DEPLOYMENT_STATUS_TYPE_OLD;

CREATE TYPE DEPLOYMENT_STATUS_TYPE AS ENUM ('ok', 'error');

ALTER TABLE DeploymentRevisionStatus
  ALTER COLUMN type TYPE DEPLOYMENT_STATUS_TYPE
    USING (type::text::DEPLOYMENT_STATUS_TYPE);

DROP TYPE DEPLOYMENT_STATUS_TYPE_OLD;
