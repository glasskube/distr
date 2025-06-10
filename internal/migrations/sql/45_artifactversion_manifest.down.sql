-- DANGER:
--   Executing this migration will invalidate all artfact versions that have
--   their manifest stored in the database. There is no fallback mechanism.
ALTER TABLE ArtifactVersion
  DROP COLUMN manifest_data;
