DROP TABLE DeploymentRevision CASCADE;
DROP TABLE DeploymentRevisionStatus;

DELETE FROM Deployment;

ALTER TABLE Deployment
  ADD COLUMN application_version_id UUID NOT NULL REFERENCES ApplicationVersion (id) ON DELETE CASCADE,
  ADD COLUMN values_yaml BYTEA;

ALTER TABLE Deployment
  DROP CONSTRAINT release_name_unique;

CREATE TABLE DeploymentStatus (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  deployment_id UUID NOT NULL REFERENCES Deployment (id) ON DELETE CASCADE,
  message TEXT NOT NULL,
  type DEPLOYMENT_STATUS_TYPE NOT NULL DEFAULT 'ok'
);
