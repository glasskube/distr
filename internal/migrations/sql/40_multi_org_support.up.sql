ALTER TABLE Organization_UserAccount ADD COLUMN created_at TIMESTAMP DEFAULT current_timestamp;

UPDATE Organization_UserAccount oua
SET created_at = ua.created_at
FROM UserAccount ua
WHERE ua.id = oua.user_account_id;
