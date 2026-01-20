CREATE TABLE DeploymentTargetLogRecord (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  deployment_target_id UUID NOT NULL REFERENCES DeploymentTarget(id) ON DELETE CASCADE,
  timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
  severity TEXT NOT NULL,
  body TEXT NOT NULL
);

CREATE INDEX deploymenttargetlogrecord_deployment_target_id_timestamp ON DeploymentTargetLogRecord(deployment_target_id, timestamp DESC);
