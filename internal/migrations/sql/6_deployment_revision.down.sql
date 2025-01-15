-- TODO

DROP TABLE IF EXISTS DeploymentRevision CASCADE;

-- TODO ??
DELETE FROM Deployment;

ALTER TABLE Deployment
  ADD COLUMN application_version_id UUID NOT NULL REFERENCES ApplicationVersion (id) ON DELETE CASCADE,
  ADD COLUMN values_yaml BYTEA;

ALTER TABLE Deployment
  DROP CONSTRAINT release_name_unique;
