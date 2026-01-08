package types

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type Deployment struct {
	Base
	DeploymentTargetID   uuid.UUID   `db:"deployment_target_id" json:"deploymentTargetId"`
	ReleaseName          *string     `db:"release_name" json:"releaseName,omitempty"`
	ApplicationLicenseID *uuid.UUID  `db:"application_license_id" json:"applicationLicenseId,omitempty"`
	DockerType           *DockerType `db:"docker_type" json:"dockerType,omitempty"`
	LogsEnabled          bool        `db:"logs_enabled" json:"logsEnabled"`
}

type DeploymentWithLatestRevision struct {
	Deployment
	DeploymentRevisionID        uuid.UUID                 `db:"deployment_revision_id" json:"deploymentRevisionId"`
	DeploymentRevisionCreatedAt time.Time                 `db:"deployment_revision_created_at" json:"deploymentRevisionCreatedAt"` //nolint:lll
	ApplicationID               uuid.UUID                 `db:"application_id" json:"applicationId"`
	ApplicationName             string                    `db:"application_name" json:"applicationName"`
	ApplicationVersionID        uuid.UUID                 `db:"application_version_id" json:"applicationVersionId"`
	ApplicationVersionName      string                    `db:"application_version_name" json:"applicationVersionName"`
	ApplicationLinkTemplate     string                    `db:"application_link_template" json:"-"`
	ApplicationLink             string                    `db:"-" json:"applicationLink"`
	ValuesYaml                  []byte                    `db:"values_yaml" json:"valuesYaml,omitempty"`
	EnvFileData                 []byte                    `db:"env_file_data" json:"envFileData,omitempty"`
	LatestStatus                *DeploymentRevisionStatus `db:"latest_status" json:"latestStatus,omitempty"`
	ForceRestart                bool                      `db:"force_restart" json:"forceRestart"`
}

func (d DeploymentWithLatestRevision) ParsedValuesFile() (result map[string]any, err error) {
	if d.ValuesYaml != nil {
		if err = yaml.Unmarshal(d.ValuesYaml, &result); err != nil {
			err = fmt.Errorf("cannot parse Deployment values file: %w", err)
		}
	}
	return result, err
}
