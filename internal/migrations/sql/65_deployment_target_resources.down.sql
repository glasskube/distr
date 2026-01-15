ALTER TABLE DeploymentTarget
  DROP COLUMN resources_cpu_request,
  DROP COLUMN resources_memory_request,
  DROP COLUMN resources_cpu_limit,
  DROP COLUMN resources_memory_limit;
