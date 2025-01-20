package types

import "time"

type DeploymentType string
type UserRole string
type HelmChartType string
type DeploymentStatusType string
type DeploymentTargetScope string

const (
	DeploymentTypeDocker     DeploymentType = "docker"
	DepolymentTypeKubernetes DeploymentType = "kubernetes"

	UserRoleVendor   UserRole = "vendor"
	UserRoleCustomer UserRole = "customer"

	HelmChartTypeRepository HelmChartType = "repository"
	HelmChartTypeOCI        HelmChartType = "oci"

	DeploymentStatusTypeOK    DeploymentStatusType = "ok"
	DeploymentStatusTypeError DeploymentStatusType = "error"

	DeploymentTargetScopeCluster   DeploymentTargetScope = "cluster"
	DeploymentTargetScopeNamespace DeploymentTargetScope = "namespace"
)

type Base struct {
	ID        string    `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type Geolocation struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}
