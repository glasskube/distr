ALTER TABLE ApplicationVersion
  ADD CONSTRAINT name_unique UNIQUE (application_id, name);
