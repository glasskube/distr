-- this migration assumes that every user is part of exactly one organization
-- if there is a user without an organization, this will fail. a manual fix would be to delete the corresponding UserAccount_TutorialProgress

ALTER TABLE UserAccount_TutorialProgress ADD COLUMN organization_id UUID REFERENCES Organization(id) ON DELETE CASCADE;

ALTER TABLE UserAccount_TutorialProgress DROP CONSTRAINT useraccount_tutorialprogress_pkey;

UPDATE UserAccount_TutorialProgress ut
SET organization_id = oua.organization_id
FROM Organization_UserAccount oua
WHERE oua.user_account_id = ut.useraccount_id;

ALTER TABLE UserAccount_TutorialProgress ALTER COLUMN organization_id SET NOT NULL;
ALTER TABLE UserAccount_TutorialProgress ADD CONSTRAINT useraccount_tutorialprogress_pkey PRIMARY KEY (useraccount_id, tutorial, organization_id);

CREATE INDEX IF NOT EXISTS fk_UserAccount_TutorialProgress_organization_id ON UserAccount_TutorialProgress (organization_id);
