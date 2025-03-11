-- TODO organization slug

-- artifact core

CREATE TABLE Artifact (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  name TEXT NOT NULL,
  organization_id UUID NOT NULL REFERENCES Organization (id) ON DELETE RESTRICT,
  CONSTRAINT Artifact_unique_name UNIQUE (organization_id, name)
);

CREATE TABLE ArtifactVersion (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  created_by_useraccount_id UUID REFERENCES UserAccount (id) ON DELETE SET NULL,
  updated_at TIMESTAMP DEFAULT current_timestamp,
  updated_by_useraccount_id UUID REFERENCES UserAccount (id) ON DELETE SET NULL,
  name TEXT NOT NULL, --- either a digest ("sha256:...") or a tag
  manifest_blob_digest TEXT NOT NULL, --- "sha256:..."
  manifest_content_type TEXT NOT NULL,
  artifact_id UUID NOT NULL REFERENCES Artifact (id) ON DELETE RESTRICT,
  CONSTRAINT ArtifactVersion_unique_name UNIQUE (artifact_id, name)
);

CREATE TABLE ArtifactVersionPart (
  artifact_version_id UUID NOT NULL REFERENCES ArtifactVersion (id) ON DELETE CASCADE,
  artifact_blob_digest TEXT NOT NULL, --- "sha256:..."
  PRIMARY KEY (artifact_version_id, artifact_blob_digest)
);

CREATE INDEX fk_Artifact_organization_id ON Artifact (organization_id);
CREATE INDEX fk_ArtifactVersion_artifact_id ON ArtifactVersion (artifact_id);

CREATE INDEX fk_ArtifactVersionPart_artifact_version_id ON ArtifactVersionPart (artifact_version_id);

-- artifact licensing
CREATE TABLE ArtifactLicense (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  name TEXT NOT NULL,
  owner_useraccount_id UUID REFERENCES UserAccount (id)
);

CREATE TABLE ArtifactLicense_Artifact (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  artifact_license_id UUID NOT NULL REFERENCES ArtifactLicense (id) ON DELETE CASCADE,
  artifact_id UUID NOT NULL REFERENCES Artifact (id) ON DELETE CASCADE,
  artifact_version_id UUID REFERENCES ArtifactVersion (id) ON DELETE CASCADE, -- NULL means all of the given artifact_id
  CONSTRAINT ArtifactLicense_Artifact_unique UNIQUE NULLS NOT DISTINCT (artifact_license_id, artifact_id, artifact_version_id)
);

CREATE INDEX IF NOT EXISTS fk_ArtifactLicense_owner_useraccount_id ON ArtifactLicense (owner_useraccount_id);
CREATE INDEX IF NOT EXISTS fk_ArtifactLicense_Artifact_artifact_license_id ON ArtifactLicense (id);

CREATE INDEX IF NOT EXISTS fk_ArtifactLicense_Artifact_artifact_license_id ON ArtifactLicense_Artifact (artifact_license_id);
CREATE INDEX IF NOT EXISTS fk_ArtifactLicense_Artifact_artifact_id ON ArtifactLicense_Artifact (artifact_id);
CREATE INDEX IF NOT EXISTS fk_ArtifactLicense_Artifact_artifact_version_id ON ArtifactLicense_Artifact (artifact_version_id);
