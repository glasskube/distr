ALTER TABLE UserAccount
  DROP COLUMN logo;
ALTER TABLE UserAccount
  DROP COLUMN logo_file_name;
ALTER TABLE UserAccount
  DROP COLUMN logo_content_type;

ALTER TABLE Application
  DROP COLUMN logo;
ALTER TABLE Application
  DROP COLUMN logo_file_name;
ALTER TABLE Application
  DROP COLUMN logo_content_type;
