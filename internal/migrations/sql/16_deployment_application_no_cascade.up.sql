ALTER TABLE DeploymentRevision
  DROP CONSTRAINT deploymentrevision_application_version_id_fkey,
  ADD CONSTRAINT deploymentrevision_application_version_id_fkey
    FOREIGN KEY (application_version_id) REFERENCES applicationversion(id) ON DELETE RESTRICT;
