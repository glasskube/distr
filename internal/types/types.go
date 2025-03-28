package types

import (
	"database/sql"
	"errors"
	"time"

	"github.com/glasskube/distr/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opencontainers/go-digest"
)

type DeploymentType string
type UserRole string
type HelmChartType string
type DeploymentStatusType string
type DeploymentTargetScope string
type Feature string

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

	FeatureLicensing Feature = "licensing"
	FeatureRegistry  Feature = "registry"
)

type Base struct {
	ID        uuid.UUID `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type Geolocation struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Digest digest.Digest

var _ sql.Scanner = util.PtrTo(Digest(""))
var _ pgtype.TextValuer = util.PtrTo(Digest(""))

func (target *Digest) Scan(src any) error {
	if srcStr, ok := src.(string); !ok {
		return errors.New("src must be a string")
	} else if h, err := digest.Parse(srcStr); err != nil {
		return err
	} else {
		*target = Digest(h)
		return nil
	}
}

// TextValue implements pgtype.TextValuer.
func (src Digest) TextValue() (pgtype.Text, error) {
	return pgtype.Text{String: string(src), Valid: true}, nil
}
