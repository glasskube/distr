package api

import "github.com/glasskube/distr/internal/types"

type AgentResource struct {
	Version types.AgentVersion `json:"version"`
}

type AgentDeployment struct {
	RevisionID string `json:"revisionId"`
}

type DockerAgentResource struct {
	AgentResource
	Deployment *DockerAgentDeployment `json:"deployment"`
}

type DockerAgentDeployment struct {
	AgentDeployment
	ComposeFile []byte `json:"composeFile"`
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
	RevisionID string                     `json:"revisionId"`
	Type       types.DeploymentStatusType `json:"type"`
	Message    string                     `json:"message"`
}
