CREATE TABLE DeploymentLogRecord (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  deployment_id UUID NOT NULL REFERENCES Deployment(id) ON DELETE CASCADE,
  deployment_revision_id UUID NOT NULL REFERENCES DeploymentRevision(id) ON DELETE CASCADE,
  resource TEXT,
  timestamp TIMESTAMP,
  severity TEXT,
  body TEXT
);

CREATE INDEX fk_DeploymentLogRecord_deployment_id ON DeploymentLogRecord (deployment_id);
CREATE INDEX fk_DeploymentLogRecord_deployment_revision_id ON DeploymentLogRecord (deployment_revision_id);
CREATE INDEX DeploymentLogRecord_created_at ON DeploymentLogRecord (deployment_id, created_at);
