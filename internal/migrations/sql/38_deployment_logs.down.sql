DROP TABLE DeploymentLogRecord;

ALTER TABLE Deployment
  DROP COLUMN logs_enabled;
