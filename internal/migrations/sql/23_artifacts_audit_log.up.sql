CREATE TABLE ArtifactVersionPull (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  artifact_version_id UUID NOT NULL REFERENCES ArtifactVersion (id) ON DELETE CASCADE,
  useraccount_id UUID REFERENCES UserAccount (id) ON DELETE SET NULL
);

CREATE INDEX fk_ArtifactVersionPull_artifact_version_id ON ArtifactVersionPull (artifact_version_id);
