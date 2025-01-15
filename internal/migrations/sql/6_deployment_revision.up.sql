-- NOTE: this migration drops all deployments and corresponding statuses that have existed, this cannot be undone!

CREATE TABLE DeploymentRevision (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP DEFAULT current_timestamp,
    deployment_id UUID NOT NULL REFERENCES Deployment (id) ON DELETE CASCADE,
    application_version_id UUID NOT NULL REFERENCES ApplicationVersion (id) ON DELETE CASCADE,
    values_yaml BYTEA
);

CREATE INDEX fk_DeploymentRevision_deployment_id ON DeploymentRevision (deployment_id);
CREATE INDEX fk_DeploymentRevision_application_version_id ON DeploymentRevision (application_version_id);

CREATE TABLE DeploymentRevisionStatus (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP DEFAULT current_timestamp,
    deployment_revision_id UUID NOT NULL REFERENCES DeploymentRevision (id) ON DELETE CASCADE,
    type DEPLOYMENT_STATUS_TYPE NOT NULL,
    message TEXT NOT NULL
);

CREATE INDEX fk_DeploymentRevisionStatus_deployment_revision_id ON DeploymentRevisionStatus (deployment_revision_id);

DELETE FROM Deployment;

DROP TABLE IF EXISTS DeploymentStatus;

ALTER TABLE Deployment
  DROP COLUMN application_version_id,
  DROP COLUMN values_yaml;
-- release_name stays in deployment and the constraint can be added again
ALTER TABLE Deployment ADD CONSTRAINT release_name_unique UNIQUE (deployment_target_id, release_name);
