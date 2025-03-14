ALTER TABLE ArtifactLicense
  DROP COLUMN organization_id,
  DROP CONSTRAINT ArtifactLicense_name_unique;
