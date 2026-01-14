ALTER TABLE DeploymentRevision
  ADD COLUMN ignore_revision_skew BOOLEAN NOT NULL DEFAULT false;
