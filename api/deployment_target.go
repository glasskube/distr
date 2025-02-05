package api

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type DeploymentTargetAccessTokenResponse struct {
	ConnectUrl   string `json:"connectUrl"`
	TargetId     string `json:"targetId"`
	TargetSecret string `json:"targetSecret"`
}

type DeploymentRequest struct {
	ID                   string  `json:"deploymentId"`
	DeploymentTargetId   string  `json:"deploymentTargetId"`
	ApplicationVersionId string  `json:"applicationVersionId"`
	ReleaseName          *string `json:"releaseName"`
	ValuesYaml           []byte  `json:"valuesYaml"`
	EnvFileData          []byte  `json:"envFileData"`
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
