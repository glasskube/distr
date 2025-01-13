CREATE TABLE IF NOT EXISTS OrganizationBranding
(
  id                         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at                 TIMESTAMP        DEFAULT current_timestamp,
  organization_id            UUID UNIQUE NOT NULL REFERENCES Organization (id) ON DELETE CASCADE,
  updated_at                 TIMESTAMP        DEFAULT current_timestamp,
  updated_by_user_account_id UUID        REFERENCES UserAccount (id) ON DELETE SET NULL,
  title                      TEXT,
  description                TEXT,
  logo                       BYTEA,
  logo_file_name             TEXT,
  logo_content_type          TEXT
);

CREATE INDEX IF NOT EXISTS fk_OrganizationBranding_organization_id ON OrganizationBranding (organization_id);
