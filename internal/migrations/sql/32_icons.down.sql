ALTER TABLE UserAccount
  DROP COLUMN icon;
ALTER TABLE UserAccount
  DROP COLUMN icon_file_name;
ALTER TABLE UserAccount
  DROP COLUMN icon_content_type;

ALTER TABLE Application
  DROP COLUMN icon;
ALTER TABLE Application
  DROP COLUMN icon_file_name;
ALTER TABLE Application
  DROP COLUMN icon_content_type;
