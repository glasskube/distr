-- IMPORTANT: Since created_by_user_account_id is not nullable, this migration is not reversible if there are existing deployment targets.

ALTER TABLE DeploymentTarget DROP COLUMN created_by_user_account_id;
