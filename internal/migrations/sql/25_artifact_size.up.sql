ALTER TABLE ArtifactVersion ADD COLUMN manifest_blob_size BIGINT NOT NULL DEFAULT 0;
ALTER TABLE ArtifactVersionPart ADD COLUMN artifact_blob_size BIGINT NOT NULL DEFAULT 0;
ALTER TABLE ArtifactVersion ALTER COLUMN manifest_blob_size DROP DEFAULT;
ALTER TABLE ArtifactVersionPart ALTER COLUMN artifact_blob_size DROP DEFAULT;
