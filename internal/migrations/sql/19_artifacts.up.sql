CREATE TABLE Artifact (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  name TEXT NOT NULL,
  organization_id UUID NOT NULL REFERENCES Organization (id) ON DELETE RESTRICT
);

CREATE TABLE ArtifactVersion (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  created_by_user_account_id UUID REFERENCES UserAccount (id) ON DELETE SET NULL,
  name TEXT NOT NULL,
  artifact_id UUID NOT NULL REFERENCES Artifact (id) ON DELETE RESTRICT
);

CREATE TABLE ArtifactBlob (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  name TEXT NOT NULL,
  is_lead BOOLEAN NOT NULL
);

CREATE TABLE ArtifactVersionPart (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  created_by_user_account_id UUID REFERENCES UserAccount (id) ON DELETE SET NULL,
  hash_md5 TEXT NOT NULL,
  hash_sha1 TEXT NOT NULL,
  hash_sha256 TEXT NOT NULL,
  hash_sha512 TEXT NOT NULL,
  artifact_version_id UUID NOT NULL REFERENCES ArtifactVersion (id) ON DELETE CASCADE,
  artifact_blob_id UUID NOT NULL REFERENCES ArtifactBlob (id) ON DELETE CASCADE
);

CREATE INDEX fk_Artifact_organization_id ON Artifact (organization_id);
CREATE INDEX fk_ArtifactVersion_artifact_id ON ArtifactVersion (artifact_id);

CREATE INDEX fk_ArtifactVersionPart_artifact_version_id ON ArtifactVersionPart (artifact_version_id);
CREATE INDEX fk_ArtifactVersionPart_artifact_blob_id ON ArtifactVersionPart (artifact_blob_id);

/*
CREATE TABLE ArtifactLicense (
                               id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                               created_at TIMESTAMP DEFAULT current_timestamp,
                               name TEXT NOT NULL,
                               owner_useraccount_id UUID REFERENCES UserAccount (id)
);

CREATE TABLE ArtifactLicense_Artifact (
                                        artifact_license_id UUID NOT NULL REFERENCES ArtifactLicense (id) ON DELETE CASCADE,
                                        artifact_id UUID NOT NULL REFERENCES Artifact (id) ON DELETE CASCADE,
                                        PRIMARY KEY (artifact_license_id, artifact_id)
);

CREATE TABLE ArtifactLicense_ArtifactTag (
                                           artifact_license_id UUID NOT NULL REFERENCES ArtifactLicense (id) ON DELETE CASCADE,
                                           artifact_id UUID NOT NULL REFERENCES Artifact (id) ON DELETE CASCADE,
                                           artifact_tag_id UUID NOT NULL REFERENCES ArtifactTag (id) ON DELETE CASCADE,
                                           PRIMARY KEY (artifact_license_id, artifact_id, artifact_tag_id)
);

CREATE INDEX fk_ArtifactLicense_owner_useraccount_id ON ArtifactLicense (owner_useraccount_id);

CREATE INDEX fk_ArtifactLicense_Artifact_artifact_license_id ON ArtifactLicense (id);
CREATE INDEX fk_ArtifactLicense_Artifact_artifact_id ON Artifact (id);

CREATE INDEX fk_ArtifactLicense_ArtifactTag_artifact_license_id ON ArtifactLicense (id);
CREATE INDEX fk_ArtifactLicense_ArtifactTag_artifact_id ON Artifact (id);
CREATE INDEX fk_ArtifactLicense_ArtifactTag_artifact_tag_id ON ArtifactTag (id);
*/
