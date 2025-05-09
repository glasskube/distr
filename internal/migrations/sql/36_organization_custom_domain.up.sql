ALTER TABLE Organization
  ADD COLUMN app_domain TEXT DEFAULT NULL,
  ADD COLUMN registry_domain TEXT DEFAULT NULL,
  ADD COLUMN email_from_address TEXT DEFAULT NULL;
