ALTER TABLE Organization
  ADD COLUMN connect_script_is_sudo BOOLEAN NOT NULL DEFAULT FALSE;
