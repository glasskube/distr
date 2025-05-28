ALTER TABLE File ALTER COLUMN organization_id DROP NOT NULL;

UPDATE File f
SET organization_id = NULL
FROM UserAccount u
WHERE u.image_id = f.id;
