ALTER TABLE Deployment
  ADD CONSTRAINT release_name_unique UNIQUE (deployment_target_id, release_name);
