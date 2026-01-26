ALTER TABLE ApplicationLicense
  DROP CONSTRAINT applicationlicense_customer_organization_id_fkey,
  ADD CONSTRAINT applicationlicense_customer_organization_id_fkey
    FOREIGN KEY (customer_organization_id)
    REFERENCES CustomerOrganization(id)
    ON DELETE CASCADE;

ALTER TABLE ArtifactLicense
  DROP CONSTRAINT artifactlicense_customer_organization_id_fkey,
  ADD CONSTRAINT artifactlicense_customer_organization_id_fkey
    FOREIGN KEY (customer_organization_id)
    REFERENCES CustomerOrganization(id)
    ON DELETE CASCADE;

ALTER TABLE Organization_UserAccount
  DROP CONSTRAINT organization_useraccount_customer_organization_id_fkey,
  ADD CONSTRAINT organization_useraccount_customer_organization_id_fkey
    FOREIGN KEY (customer_organization_id)
    REFERENCES CustomerOrganization(id)
    ON DELETE CASCADE;

ALTER TABLE DeploymentTarget
  DROP CONSTRAINT deploymenttarget_customer_organization_id_fkey,
  ADD CONSTRAINT deploymenttarget_customer_organization_id_fkey
    FOREIGN KEY (customer_organization_id)
    REFERENCES CustomerOrganization(id)
    ON DELETE CASCADE;

ALTER TABLE Deployment
  DROP CONSTRAINT deployment_application_license_id_fkey,
  ADD CONSTRAINT deployment_application_license_id_fkey
    FOREIGN KEY (application_license_id)
    REFERENCES ApplicationLicense(id)
    DEFERRABLE INITIALLY IMMEDIATE;
