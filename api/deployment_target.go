package api

import (
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
)

type DeploymentTargetAccessTokenResponse struct {
	ConnectURL     string    `json:"connectUrl"`
	TargetID       uuid.UUID `json:"targetId"`
	TargetSecret   string    `json:"targetSecret"`
	ConnectCommand string    `json:"connectCommand"`
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
	IgnoreRevisionSkew   bool              `json:"ignoreRevisionSkew"`
}

func (d *DeploymentRequest) GetValuesYAML() []byte {
	return d.ValuesYaml
}

func (d *DeploymentRequest) GetEnvFileData() []byte {
	return d.EnvFileData
}

type PatchDeploymentRequest struct {
	LogsEnabled *bool `json:"logsEnabled,omitempty"`
}
