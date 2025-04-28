package types

import (
	"fmt"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type Deployment struct {
	Base
	DeploymentTargetID   uuid.UUID   `db:"deployment_target_id" json:"deploymentTargetId"`
	ReleaseName          *string     `db:"release_name" json:"releaseName,omitempty"`
	ApplicationLicenseID *uuid.UUID  `db:"application_license_id" json:"applicationLicenseId,omitempty"`
	DockerType           *DockerType `db:"docker_type" json:"dockerType,omitempty"`
}

type DeploymentWithLatestRevision struct {
	Deployment
	DeploymentRevisionID   uuid.UUID                 `db:"deployment_revision_id" json:"deploymentRevisionId"`
	ApplicationID          uuid.UUID                 `db:"application_id" json:"applicationId"`
	ApplicationName        string                    `db:"application_name" json:"applicationName"`
	ApplicationVersionID   uuid.UUID                 `db:"application_version_id" json:"applicationVersionId"`
	ApplicationVersionName string                    `db:"application_version_name" json:"applicationVersionName"`
	ValuesYaml             []byte                    `db:"values_yaml" json:"valuesYaml,omitempty"`
	EnvFileData            []byte                    `db:"env_file_data" json:"envFileData,omitempty"`
	LatestStatus           *DeploymentRevisionStatus `db:"latest_status" json:"latestStatus,omitempty"`
}

func (d DeploymentWithLatestRevision) ParsedValuesFile() (result map[string]any, err error) {
	if d.ValuesYaml != nil {
		if err = yaml.Unmarshal(d.ValuesYaml, &result); err != nil {
			err = fmt.Errorf("cannot parse Deployment values file: %w", err)
		}
	}
	return
}
