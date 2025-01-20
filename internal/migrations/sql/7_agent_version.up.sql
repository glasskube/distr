CREATE TABLE AgentVersion (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  name TEXT NOT NULL UNIQUE,
  manifest_file_revision TEXT NOT NULL,
  compose_file_revision TEXT NOT NULL
);

CREATE INDEX AgentVersion_name ON AgentVersion (name);

ALTER TABLE DeploymentTarget
  ADD COLUMN agent_version_id UUID REFERENCES AgentVersion (id),
  ADD COLUMN reported_agent_version_id UUID REFERENCES AgentVersion (id) ON DELETE SET NULL;

CREATE INDEX fk_DeploymentTarget_agent_version_id ON DeploymentTarget (agent_version_id);

WITH inserted AS (
  INSERT INTO AgentVersion (name, manifest_file_revision, compose_file_revision)
    VALUES ('0.8.2', 'v1', 'v1')
    RETURNING id
)
UPDATE DeploymentTarget
SET agent_version_id = inserted.id
FROM inserted;

ALTER TABLE DeploymentTarget
  ALTER COLUMN agent_version_id SET NOT NULL;
