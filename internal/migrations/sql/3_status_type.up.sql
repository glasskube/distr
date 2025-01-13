CREATE TYPE DEPLOYMENT_STATUS_TYPE AS ENUM ('ok', 'error'); -- TODO maybe an 'UNKNOWN' type?

-- TODO for DeploymentStatus table too probably

ALTER TABLE DeploymentStatus ADD COLUMN type DEPLOYMENT_STATUS_TYPE NOT NULL DEFAULT 'ok';
