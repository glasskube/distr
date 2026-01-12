CREATE TABLE Secret (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT now(),
  updated_at TIMESTAMP DEFAULT now(),
  updated_by_useraccount_id UUID REFERENCES UserAccount(id) ON DELETE SET NULL,
  key TEXT NOT NULL,
  value TEXT NOT NULL,
  organization_id UUID NOT NULL REFERENCES Organization(id) ON DELETE CASCADE,
  customer_organization_id UUID REFERENCES CustomerOrganization(id) ON DELETE CASCADE,
  CONSTRAINT unique_secret_key UNIQUE NULLS NOT DISTINCT (key, organization_id, customer_organization_id)
);

CREATE INDEX idx_secret_organization_id_customer_organization_id ON Secret(organization_id, customer_organization_id);
