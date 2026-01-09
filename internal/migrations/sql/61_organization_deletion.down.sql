ALTER TABLE Artifact
DROP CONSTRAINT artifact_organization_id_fkey,
ADD CONSTRAINT artifact_organization_id_fkey
  FOREIGN KEY (organization_id) REFERENCES Organization(id);

ALTER TABLE ArtifactLicense
DROP CONSTRAINT artifactlicense_organization_id_fkey,
ADD CONSTRAINT artifactlicense_organization_id_fkey
  FOREIGN KEY (organization_id) REFERENCES Organization(id);

ALTER TABLE CustomerOrganization
DROP CONSTRAINT customerorganization_organization_id_fkey,
ADD CONSTRAINT customerorganization_organization_id_fkey
  FOREIGN KEY (organization_id) REFERENCES Organization(id);

ALTER TABLE DeploymentRevision
DROP CONSTRAINT deploymentrevision_application_version_id_fkey,
ADD CONSTRAINT deploymentrevision_application_version_id_fkey
  FOREIGN KEY (application_version_id) REFERENCES ApplicationVersion(id);

ALTER TABLE Organization
  DROP COLUMN deleted_at;
