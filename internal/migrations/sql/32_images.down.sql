ALTER TABLE UserAccount
  DROP COLUMN image;
ALTER TABLE UserAccount
  DROP COLUMN image_file_name;
ALTER TABLE UserAccount
  DROP COLUMN image_content_type;

ALTER TABLE Application
  DROP COLUMN image;
ALTER TABLE Application
  DROP COLUMN image_file_name;
ALTER TABLE Application
  DROP COLUMN image_content_type;

ALTER TABLE Artifact
  DROP COLUMN image;
ALTER TABLE Artifact
  DROP COLUMN image_file_name;
ALTER TABLE Artifact
  DROP COLUMN image_content_type;
