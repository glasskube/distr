ALTER TABLE DeploymentTarget
  ADD COLUMN resources_cpu_request text,
  ADD COLUMN resources_memory_request text,
  ADD COLUMN resources_cpu_limit text,
  ADD COLUMN resources_memory_limit text,
  ADD CONSTRAINT resources_check CHECK (
    (resources_cpu_limit IS NULL AND resources_memory_limit IS NULL AND resources_cpu_request IS NULL AND resources_memory_request IS NULL)
    OR
    (resources_cpu_limit IS NOT NULL AND resources_memory_limit IS NOT NULL AND resources_cpu_request IS NOT NULL AND resources_memory_request IS NOT NULL)
  ),
  ADD CONSTRAINT type_docker_resources_check CHECK (
    type != 'docker'
    OR
    (resources_cpu_limit IS NULL AND resources_memory_limit IS NULL AND resources_cpu_request IS NULL AND resources_memory_request IS NULL)
  );
