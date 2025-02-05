package types

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Deployment struct {
	Base
	DeploymentTargetId string  `db:"deployment_target_id" json:"deploymentTargetId"`
	ReleaseName        *string `db:"release_name" json:"releaseName,omitempty"`
}

type DeploymentWithLatestRevision struct {
	Deployment
	DeploymentRevisionID   string                    `db:"deployment_revision_id" json:"deploymentRevisionId"`
	ApplicationId          string                    `db:"application_id" json:"applicationId"`
	ApplicationName        string                    `db:"application_name" json:"applicationName"`
	ApplicationVersionId   string                    `db:"application_version_id" json:"applicationVersionId"`
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
