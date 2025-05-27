package types

import (
	"database/sql"
	"errors"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type (
	DeploymentType        string
	UserRole              string
	HelmChartType         string
	DeploymentStatusType  string
	DeploymentTargetScope string
	Feature               string
	DockerType            string
	Tutorial              string
	FileScope             string
)

const (
	DeploymentTypeDocker     DeploymentType = "docker"
	DepolymentTypeKubernetes DeploymentType = "kubernetes"

	UserRoleVendor   UserRole = "vendor"
	UserRoleCustomer UserRole = "customer"

	HelmChartTypeRepository HelmChartType = "repository"
	HelmChartTypeOCI        HelmChartType = "oci"

	DockerTypeCompose DockerType = "compose"
	DockerTypeSwarm   DockerType = "swarm"

	DeploymentStatusTypeOK          DeploymentStatusType = "ok"
	DeploymentStatusTypeProgressing DeploymentStatusType = "progressing"
	DeploymentStatusTypeError       DeploymentStatusType = "error"

	DeploymentTargetScopeCluster   DeploymentTargetScope = "cluster"
	DeploymentTargetScopeNamespace DeploymentTargetScope = "namespace"

	FeatureLicensing Feature = "licensing"

	TutorialBranding Tutorial = "branding"
	TutorialAgents   Tutorial = "agents"
	TutorialRegistry Tutorial = "registry"

	FileScopePlatform     FileScope = "platform"
	FileScopeOrganization FileScope = "organization"
)

type Base struct {
	ID        uuid.UUID `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type Geolocation struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Image struct {
	Image            []byte  `db:"image" json:"image"`
	ImageFileName    *string `db:"image_file_name" json:"imageFileName"`
	ImageContentType *string `db:"image_content_type" json:"imageContentType"`
}

type Digest v1.Hash

var (
	_ sql.Scanner       = &Digest{}
	_ pgtype.TextValuer = &Digest{}
)

func (target *Digest) Scan(src any) error {
	if srcStr, ok := src.(string); !ok {
		return errors.New("src must be a string")
	} else if h, err := v1.NewHash(srcStr); err != nil {
		return err
	} else {
		*target = Digest(h)
		return nil
	}
}

// TextValue implements pgtype.TextValuer.
func (src Digest) TextValue() (pgtype.Text, error) {
	return pgtype.Text{String: v1.Hash(src).String(), Valid: true}, nil
}

func (h Digest) MarshalJSON() ([]byte, error) {
	return v1.Hash(h).MarshalJSON()
}
