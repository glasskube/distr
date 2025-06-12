package types

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/glasskube/distr/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opencontainers/go-digest"
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
	OIDCProvider          string
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

	OIDCProviderGithub    OIDCProvider = "github"
	OIDCProviderGoogle    OIDCProvider = "google"
	OIDCProviderMicrosoft OIDCProvider = "microsoft"
)

type Base struct {
	ID        uuid.UUID `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type Image struct {
	Image            []byte  `db:"image" json:"image"`
	ImageFileName    *string `db:"image_file_name" json:"imageFileName"`
	ImageContentType *string `db:"image_content_type" json:"imageContentType"`
}

type Digest digest.Digest

var (
	_ sql.Scanner       = util.PtrTo(Digest(""))
	_ pgtype.TextValuer = util.PtrTo(Digest(""))
)

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

func (h Digest) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(h))
}
