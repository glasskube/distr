package api

import "github.com/glasskube/distr/internal/types"

type AgentResource struct {
	RevisionID string `json:"revisionId"`
}

type DockerAgentResource struct {
	AgentResource
	ComposeFile []byte `json:"composeFile"`
}

type KubernetesAgentResource struct {
	AgentResource
	Namespace  string                     `json:"namespace"`
	Deployment *KubernetesAgentDeployment `json:"deployment"`
	Version    types.AgentVersion         `json:"version"`
}

type KubernetesAgentDeployment struct {
	RevisionID   string         `json:"revisionId"`
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
