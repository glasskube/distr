-- this migration assumes that every user is part of exactly one organization
-- if there is a user without an organization, this will fail. a manual fix would be to delete the corresponding AccessToken

ALTER TABLE AccessToken ADD COLUMN organization_id UUID REFERENCES Organization(id) ON DELETE CASCADE;

UPDATE AccessToken at
SET organization_id = oua.organization_id
FROM Organization_UserAccount oua
WHERE oua.user_account_id = at.user_account_id;

ALTER TABLE AccessToken ALTER COLUMN organization_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS fk_AccessToken_organization_id ON AccessToken (organization_id);
