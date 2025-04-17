CREATE TABLE IF NOT EXISTS File
(
  id              UUID PRIMARY KEY   DEFAULT gen_random_uuid(),
  created_at      TIMESTAMP NOT NULL DEFAULT current_timestamp,
  organization_id UUID      NOT NULL REFERENCES Organization (id) ON DELETE CASCADE,
  content_type    TEXT      NOT NULL,
  data            BYTEA     NOT NULL,
  file_name       TEXT      NOT NULL,
  file_size       INT       NOT NULL
);

CREATE INDEX IF NOT EXISTS fk_File_organization_id ON File (organization_id);


ALTER TABLE UserAccount
  ADD COLUMN image_id UUID DEFAULT NULL REFERENCES File (id) ON DELETE SET NULL;

ALTER TABLE Application
  ADD COLUMN image_id UUID DEFAULT NULL REFERENCES File (id) ON DELETE SET NULL;

ALTER TABLE Artifact
  ADD COLUMN image_id UUID DEFAULT NULL REFERENCES File (id) ON DELETE SET NULL;
