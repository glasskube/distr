DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'docker_type') THEN
        CREATE TYPE DOCKER_TYPE AS ENUM ('compose', 'swarm');
    END IF;
END $$;

ALTER TABLE ApplicationVersion
  ADD COLUMN IF NOT EXISTS docker_type DOCKER_TYPE DEFAULT 'compose';

