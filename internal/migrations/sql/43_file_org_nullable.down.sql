UPDATE File f
SET organization_id = oua.organization_id
FROM Organization_UserAccount oua
INNER JOIN UserAccount u ON oua.user_account_id = u.id
WHERE u.image_id = f.id;

ALTER TABLE File ALTER COLUMN organization_id SET NOT NULL;
