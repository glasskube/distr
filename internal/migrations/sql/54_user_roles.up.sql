ALTER TABLE Organization_UserAccount DROP COLUMN user_role;

DROP TYPE USER_ROLE;

CREATE TYPE USER_ROLE AS ENUM ('read_only', 'read_write', 'admin');

ALTER TABLE Organization_UserAccount
  ADD COLUMN user_role USER_ROLE NOT NULL DEFAULT 'admin';
