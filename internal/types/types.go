package types

import "time"

type DeploymentType string
type UserRole string

const (
	DeploymentTypeDocker     = "docker"
	DepolymentTypeKubernetes = "kubernetes"
	UserRoleDistributor      = "distributor"
	UserRoleCustomer         = "customer"
)

type Base struct {
	ID        string    `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type Geolocation struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}
