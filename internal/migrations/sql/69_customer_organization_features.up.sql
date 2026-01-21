CREATE TYPE CUSTOMER_ORGANIZATION_FEATURE AS ENUM ('deployment_targets', 'artifacts');

ALTER TABLE CustomerOrganization ADD COLUMN features CUSTOMER_ORGANIZATION_FEATURE[] DEFAULT '{deployment_targets,artifacts}' NOT NULL;
