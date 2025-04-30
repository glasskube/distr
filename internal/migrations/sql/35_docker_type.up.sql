CREATE TYPE DOCKER_TYPE AS ENUM ('compose', 'swarm');

ALTER TABLE Deployment
  ADD COLUMN docker_type DOCKER_TYPE;

UPDATE Deployment d
  SET docker_type = 'compose'
FROM DeploymentTarget dt
  WHERE d.deployment_target_id = dt.id
  AND dt.type = 'docker';
