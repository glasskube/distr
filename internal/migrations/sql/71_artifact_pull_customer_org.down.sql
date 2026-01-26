DROP INDEX fk_ArtifactVersionPull_customer_organization_id;

ALTER TABLE ArtifactVersionPull
  DROP COLUMN customer_organization_id;
