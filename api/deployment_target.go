package api

import (
	"fmt"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type DeploymentTargetAccessTokenResponse struct {
	ConnectURL   string    `json:"connectUrl"`
	TargetID     uuid.UUID `json:"targetId"`
	TargetSecret string    `json:"targetSecret"`
}

type DeploymentRequest struct {
	ID                   uuid.UUID `json:"deploymentId"`
	DeploymentTargetID   uuid.UUID `json:"deploymentTargetId"`
	ApplicationVersionID uuid.UUID `json:"applicationVersionId"`
	ReleaseName          *string   `json:"releaseName"`
	ValuesYaml           []byte    `json:"valuesYaml"`
	EnvFileData          []byte    `json:"envFileData"`
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
