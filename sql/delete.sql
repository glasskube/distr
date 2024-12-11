DROP INDEX IF EXISTS fk_ApplicationVersion_application_id;

DROP INDEX IF EXISTS fk_DeploymentTargetStatus_deployment_target_id;

DROP INDEX IF EXISTS fk_Deployment_deployment_target_id;

DROP INDEX IF EXISTS fk_Deployment_application_version_id;

DROP INDEX IF EXISTS fk_DeploymentStatus_deployment_target_id;

DROP INDEX IF EXISTS fk_Organization_Application_organization_id;

DROP INDEX IF EXISTS fk_Organization_Application_application_id;

DROP INDEX IF EXISTS fk_Organization_DeploymentTarget_organization_id;

DROP INDEX IF EXISTS fk_Organization_DeploymentTarget_user_account_id;

DROP TABLE IF EXISTS Organization_UserAccount CASCADE;

DROP TABLE IF EXISTS UserAccount CASCADE;

DROP TABLE IF EXISTS DeploymentStatus CASCADE;

DROP TABLE IF EXISTS Deployment CASCADE;

DROP TABLE IF EXISTS DeploymentTargetStatus CASCADE;

DROP TABLE IF EXISTS DeploymentTarget CASCADE;

DROP TABLE IF EXISTS ApplicationVersion CASCADE;

DROP TABLE IF EXISTS Application CASCADE;

DROP TABLE IF EXISTS Organization CASCADE;

DROP TYPE IF EXISTS DEPLOYMENT_TYPE CASCADE;

DROP TYPE IF EXISTS USER_ROLE CASCADE;
