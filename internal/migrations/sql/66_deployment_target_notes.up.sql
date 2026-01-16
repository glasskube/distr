CREATE TABLE DeploymentTargetNotes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  deployment_target_id UUID REFERENCES DeploymentTarget (id) ON DELETE CASCADE,
  customer_organization_id UUID REFERENCES CustomerOrganization (id) ON DELETE CASCADE,
  updated_by_useraccount_id UUID REFERENCES UserAccount (id) ON DELETE SET NULL,
  updated_at TIMESTAMP NOT NULL,
  notes TEXT NOT NULL,
  UNIQUE NULLS NOT DISTINCT (deployment_target_id, customer_organization_id)
);

CREATE INDEX IF NOT EXISTS DeploymentTargetNotes_deployment_customer_org_key
ON DeploymentTargetNotes (deployment_target_id, customer_organization_id);
