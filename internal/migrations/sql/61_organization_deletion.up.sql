ALTER TABLE Artifact
DROP CONSTRAINT artifact_organization_id_fkey,
ADD CONSTRAINT artifact_organization_id_fkey
  FOREIGN KEY (organization_id) REFERENCES Organization(id)
  ON DELETE CASCADE;

ALTER TABLE ArtifactLicense
DROP CONSTRAINT artifactlicense_organization_id_fkey,
ADD CONSTRAINT artifactlicense_organization_id_fkey
  FOREIGN KEY (organization_id) REFERENCES Organization(id)
  ON DELETE CASCADE;

ALTER TABLE CustomerOrganization
DROP CONSTRAINT customerorganization_organization_id_fkey,
ADD CONSTRAINT customerorganization_organization_id_fkey
  FOREIGN KEY (organization_id) REFERENCES Organization(id)
  ON DELETE CASCADE;

ALTER TABLE DeploymentRevision
  DROP CONSTRAINT deploymentrevision_application_version_id_fkey,
  ADD CONSTRAINT deploymentrevision_application_version_id_fkey
    FOREIGN KEY (application_version_id) REFERENCES ApplicationVersion(id)
    DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE Organization
  ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;
