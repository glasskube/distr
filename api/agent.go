package api

import (
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
)

type AgentResource struct {
	Version        types.AgentVersion `json:"version"`
	Namespace      string             `json:"namespace,omitempty"`
	MetricsEnabled bool               `json:"metricsEnabled"`
	// Deprecated: This property will be removed in v2. Please consider using Deployments instead.
	Deployment  *AgentDeployment  `json:"deployment,omitempty"`
	Deployments []AgentDeployment `json:"deployments,omitempty"`
}

type AgentRegistryAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AgentDeployment struct {
	ID           uuid.UUID                    `json:"id"`
	RevisionID   uuid.UUID                    `json:"revisionId"`
	RegistryAuth map[string]AgentRegistryAuth `json:"registryAuth"`

	// Docker specific data

	ComposeFile []byte            `json:"composeFile"`
	EnvFile     []byte            `json:"envFile"`
	DockerType  *types.DockerType `json:"dockerType"`

	// Kubernetes specific data

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

type AgentDeploymentTargetMetrics struct {
	CPUCoresM   int64   `json:"cpuCoresM" db:"cpu_cores_m"`
	CPUUsage    float64 `json:"cpuUsage" db:"cpu_usage"`
	MemoryBytes int64   `json:"memoryBytes" db:"memory_bytes"`
	MemoryUsage float64 `json:"memoryUsage" db:"memory_usage"`
}
