DROP INDEX IF EXISTS fk_ApplicationVersion_application_id;

DROP INDEX IF EXISTS fk_DeploymentTargetStatus_deployment_target_id;

DROP INDEX IF EXISTS fk_Deployment_deployment_target_id;

DROP INDEX IF EXISTS fk_Deployment_application_version_id;

DROP INDEX IF EXISTS fk_DeploymentStatus_deployment_target_id;

DROP INDEX IF EXISTS fk_Organization_Application_organization_id;

DROP INDEX IF EXISTS fk_Organization_Application_application_id;

DROP INDEX IF EXISTS fk_Organization_DeploymentTarget_organization_id;

DROP INDEX IF EXISTS fk_Organization_DeploymentTarget_user_account_id;

DROP TABLE IF EXISTS Organization_UserAccount;

DROP TABLE IF EXISTS Organization_Application;

DROP TABLE IF EXISTS Organization_DeploymentTarget;

DROP TABLE IF EXISTS UserAccount;

DROP TABLE IF EXISTS DeploymentStatus;

DROP TABLE IF EXISTS Deployment;

DROP TABLE IF EXISTS DeploymentTargetStatus;

DROP TABLE IF EXISTS DeploymentTarget;

DROP TABLE IF EXISTS ApplicationVersion;

DROP TABLE IF EXISTS Application;

DROP TABLE IF EXISTS Organization;

DROP TYPE IF EXISTS DEPLOYMENT_TYPE;
