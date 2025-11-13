ALTER TABLE ArtifactVersion
  DROP CONSTRAINT artifactversion_artifact_id_fkey,
  ADD CONSTRAINT artifactversion_artifact_id_fkey
    FOREIGN KEY (artifact_id) REFERENCES Artifact (id) ON DELETE RESTRICT;
