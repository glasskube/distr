ALTER TABLE UserAccount
  DROP COLUMN image_id;

ALTER TABLE Application
  DROP COLUMN image_id;

ALTER TABLE Artifact
  DROP COLUMN image_id;

DROP TABLE IF EXISTS File;
