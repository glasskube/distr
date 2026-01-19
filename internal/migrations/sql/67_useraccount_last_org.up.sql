ALTER TABLE UserAccount
  ADD COLUMN last_used_organization_id UUID DEFAULT NULL REFERENCES Organization(id) ON DELETE SET NULL;
