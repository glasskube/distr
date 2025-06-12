package api

import (
	"fmt"

	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type DeploymentTargetAccessTokenResponse struct {
	ConnectURL   string    `json:"connectUrl"`
	TargetID     uuid.UUID `json:"targetId"`
	TargetSecret string    `json:"targetSecret"`
}

type DeploymentRequest struct {
	DeploymentID         *uuid.UUID        `json:"deploymentId"`
	DeploymentTargetID   uuid.UUID         `json:"deploymentTargetId"`
	ApplicationVersionID uuid.UUID         `json:"applicationVersionId"`
	ApplicationLicenseID *uuid.UUID        `json:"applicationLicenseId"`
	ReleaseName          *string           `json:"releaseName"`
	ValuesYaml           []byte            `json:"valuesYaml"`
	DockerType           *types.DockerType `json:"dockerType"`
	EnvFileData          []byte            `json:"envFileData"`
	LogsEnabled          bool              `json:"logsEnabled"`
	ForceRestart         bool              `json:"forceRestart"`
}

func (d DeploymentRequest) ParsedValuesFile() (result map[string]any, err error) {
	// TODO deduplicate
	if d.ValuesYaml != nil {
		if err = yaml.Unmarshal(d.ValuesYaml, &result); err != nil {
			err = fmt.Errorf("cannot parse Deployment values file: %w", err)
		}
	}
	return
}

type PatchDeploymentRequest struct {
	LogsEnabled *bool `json:"logsEnabled,omitempty"`
}
