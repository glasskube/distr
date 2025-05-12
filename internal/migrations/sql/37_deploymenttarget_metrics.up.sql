ALTER TABLE DeploymentTarget ADD COLUMN metrics_enabled BOOLEAN NOT NULL DEFAULT true;

UPDATE DeploymentTarget SET metrics_enabled = false WHERE scope = 'namespace';

CREATE TABLE DeploymentTargetMetrics (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
  deployment_target_id UUID NOT NULL REFERENCES DeploymentTarget(id) ON DELETE CASCADE,
  cpu_cores_m BIGINT NOT NULL,
  cpu_usage FLOAT NOT NULL,
  memory_bytes BIGINT NOT NULL,
  memory_usage FLOAT NOT NULL
);

-- TODO Probably another index depending on the query

CREATE INDEX DeploymentTargetMetrics_deployment_target_id ON DeploymentTargetMetrics(deployment_target_id);
