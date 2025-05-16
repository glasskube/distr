-- set docker type compose for accidentally created docker deployments without docker type (hello-distr-tutorial bug)

UPDATE Deployment d
SET docker_type = 'compose'
FROM DeploymentTarget dt
WHERE d.deployment_target_id = dt.id
  AND dt.type = 'docker' AND d.docker_type IS NULL;
