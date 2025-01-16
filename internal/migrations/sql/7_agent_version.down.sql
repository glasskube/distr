ALTER TABLE DeploymentTarget
  DROP COLUMN agent_version_id,
  DROP COLUMN reported_agent_version_id;

DROP TABLE AgentVersion;
