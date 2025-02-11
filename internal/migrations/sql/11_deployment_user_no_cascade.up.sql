ALTER TABLE DeploymentTarget
  DROP CONSTRAINT deploymenttarget_created_by_user_account_id_fkey,
  ADD CONSTRAINT deploymenttarget_created_by_user_account_id_fkey
    FOREIGN KEY (created_by_user_account_id) REFERENCES UserAccount (id) ON DELETE RESTRICT;
