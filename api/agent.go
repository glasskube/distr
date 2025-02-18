package api

import (
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
)

type AgentResource struct {
	Version types.AgentVersion `json:"version"`
}

type AgentRegistryAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AgentDeployment struct {
	ID           uuid.UUID                    `json:"id"`
	RevisionID   uuid.UUID                    `json:"revisionId"`
	RegistryAuth map[string]AgentRegistryAuth `json:"registryAuth"`
}

type DockerAgentResource struct {
	AgentResource
	Deployment *DockerAgentDeployment `json:"deployment"`
}

type DockerAgentDeployment struct {
	AgentDeployment
	ComposeFile []byte `json:"composeFile"`
	EnvFile     []byte `json:"envFile"`
}

type KubernetesAgentResource struct {
	AgentResource
	Namespace  string                     `json:"namespace"`
	Deployment *KubernetesAgentDeployment `json:"deployment"`
}

type KubernetesAgentDeployment struct {
	AgentDeployment
	ReleaseName  string         `json:"releaseName"`
	ChartUrl     string         `json:"chartUrl"`
	ChartName    string         `json:"chartName"`
	ChartVersion string         `json:"chartVersion"`
	Values       map[string]any `json:"values"`
}

type AgentDeploymentStatus struct {
	RevisionID uuid.UUID                  `json:"revisionId"`
	Type       types.DeploymentStatusType `json:"type"`
	Message    string                     `json:"message"`
}
