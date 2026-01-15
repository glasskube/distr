package types

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/distr-sh/distr/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opencontainers/go-digest"
)

type UserRole string

const (
	UserRoleReadOnly  UserRole = "read_only"
	UserRoleReadWrite UserRole = "read_write"
	UserRoleAdmin     UserRole = "admin"
)

func ParseUserRole(value string) (UserRole, error) {
	switch value {
	case string(UserRoleReadOnly):
		return UserRoleReadOnly, nil
	case string(UserRoleReadWrite):
		return UserRoleReadWrite, nil
	case string(UserRoleAdmin):
		return UserRoleAdmin, nil
	default:
		return "", errors.New("invalid user role")
	}
}

type SubscriptionType string

func (st SubscriptionType) IsPro() bool {
	return st == SubscriptionTypeTrial || st == SubscriptionTypePro || st == SubscriptionTypeEnterprise
}

const (
	SubscriptionTypeCommunity  SubscriptionType = "community"
	SubscriptionTypeStarter    SubscriptionType = "starter"
	SubscriptionTypePro        SubscriptionType = "pro"
	SubscriptionTypeEnterprise SubscriptionType = "enterprise"
	SubscriptionTypeTrial      SubscriptionType = "trial"
)

var NonProSubscriptionTypes = []SubscriptionType{
	SubscriptionTypeCommunity,
	SubscriptionTypeStarter,
}

var AllSubscriptionTypes = []SubscriptionType{
	SubscriptionTypeCommunity,
	SubscriptionTypeStarter,
	SubscriptionTypePro,
	SubscriptionTypeEnterprise,
	SubscriptionTypeTrial,
}

type Feature string

const (
	FeatureLicensing              Feature = "licensing"
	FeaturePrePostScripts         Feature = "pre_post_scripts"
	FeatureArtifactVersionMutable Feature = "artifact_version_mutable"
)

type (
	DeploymentType        string
	HelmChartType         string
	DeploymentStatusType  string
	DeploymentTargetScope string
	DockerType            string
	Tutorial              string
	FileScope             string
	SubscriptionPeriod    string
)

const (
	DeploymentTypeDocker     DeploymentType = "docker"
	DeploymentTypeKubernetes DeploymentType = "kubernetes"

	HelmChartTypeRepository HelmChartType = "repository"
	HelmChartTypeOCI        HelmChartType = "oci"

	DockerTypeCompose DockerType = "compose"
	DockerTypeSwarm   DockerType = "swarm"

	DeploymentStatusTypeOK          DeploymentStatusType = "ok"
	DeploymentStatusTypeProgressing DeploymentStatusType = "progressing"
	DeploymentStatusTypeError       DeploymentStatusType = "error"

	DeploymentTargetScopeCluster   DeploymentTargetScope = "cluster"
	DeploymentTargetScopeNamespace DeploymentTargetScope = "namespace"

	TutorialBranding      Tutorial  = "branding"
	TutorialAgents        Tutorial  = "agents"
	TutorialRegistry      Tutorial  = "registry"
	FileScopePlatform     FileScope = "platform"
	FileScopeOrganization FileScope = "organization"

	SubscriptionPeriodMonthly SubscriptionPeriod = "monthly"
	SubscriptionPeriodYearly  SubscriptionPeriod = "yearly"
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
