ALTER TABLE ApplicationVersion
  DROP COLUMN IF EXISTS docker_type;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'docker_type') THEN
        DROP TYPE DOCKER_TYPE;
    END IF;
END $$;
