ALTER TABLE UserAccount
  ADD COLUMN logo BYTEA DEFAULT NULL;
ALTER TABLE UserAccount
  ADD COLUMN logo_file_name TEXT DEFAULT NULL;
ALTER TABLE UserAccount
  ADD COLUMN logo_content_type TEXT DEFAULT NULL;

ALTER TABLE Application
  ADD COLUMN logo BYTEA DEFAULT NULL;
ALTER TABLE Application
  ADD COLUMN logo_file_name TEXT DEFAULT NULL;
ALTER TABLE Application
  ADD COLUMN logo_content_type TEXT DEFAULT NULL;
