CREATE INDEX DeploymentLogRecord_deployment_id_resource_timestamp
  ON DeploymentLogRecord(deployment_id, resource, timestamp DESC);

DROP INDEX DeploymentLogRecord_created_at;
