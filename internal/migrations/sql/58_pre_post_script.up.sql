ALTER TABLE Organization
  ADD COLUMN pre_connect_script TEXT,
  ADD COLUMN post_connect_script TEXT;

ALTER TYPE Feature
  ADD VALUE IF NOT EXISTS 'pre_post_scripts';
