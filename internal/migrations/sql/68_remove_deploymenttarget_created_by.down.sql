-- This will fail if there are existing deployment targets.
ALTER TABLE DeploymentTarget
  ADD COLUMN created_by_user_account_id UUID NOT NULL REFERENCES UserAccount(id) ON DELETE RESTRICT;
