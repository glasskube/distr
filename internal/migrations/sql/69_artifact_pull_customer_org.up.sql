ALTER TABLE ArtifactVersionPull
  ADD COLUMN customer_organization_id UUID REFERENCES CustomerOrganization(id) ON DELETE CASCADE;

CREATE INDEX fk_ArtifactVersionPull_customer_organization_id ON ArtifactVersionPull (customer_organization_id);
