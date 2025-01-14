package api

import "github.com/glasskube/cloud/internal/types"

type KubernetesAgentResource struct {
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
	Type    types.DeploymentStatusType `json:"type"`
	Message string                     `json:"message"`
}
