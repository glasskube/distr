CREATE TABLE AgentVersion (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  name TEXT NOT NULL UNIQUE,
  manifest_file_revision TEXT NOT NULL,
  compose_file_revision TEXT NOT NULL
);

ALTER TABLE DeploymentTarget
  ADD COLUMN agent_version_id UUID REFERENCES AgentVersion (id);
