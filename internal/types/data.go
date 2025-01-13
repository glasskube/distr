package types

import (
	"errors"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

type Application struct {
	Base
	OrganizationID string               `db:"organization_id" json:"-"`
	Name           string               `db:"name" json:"name"`
	Type           DeploymentType       `db:"type" json:"type"`
	Versions       []ApplicationVersion `db:"versions" json:"versions"`
}

type ApplicationVersion struct {
	// TODO unfortunately Base nested type doesn't work when ApplicationVersion is a nested row in an SQL query
	ID        string    `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	Name      string    `db:"name" json:"name"`

	ChartType    *HelmChartType `db:"chart_type" json:"chartType,omitempty"`
	ChartName    *string        `db:"chart_name" json:"chartName,omitempty"`
	ChartUrl     *string        `db:"chart_url" json:"chartUrl,omitempty"`
	ChartVersion *string        `db:"chart_version" json:"chartVersion,omitempty"`

	// awful but relevant: the following must be defined after the ChartType, because somehow order matters
	// for pgx at collecting the subrows (relevant at getting application + list of its versions with these
	// array aggregations) – long term it should probably be refactored because this is such a pitfall
	// https://github.com/jackc/pgx/issues/1585#issuecomment-1528810634
	ValuesFileData   []byte `db:"values_file_data" json:"-"`
	TemplateFileData []byte `db:"template_file_data" json:"-"`
	ComposeFileData  []byte `db:"compose_file_data" json:"-"`

	ApplicationId string `db:"application_id" json:"applicationId"`
}

func (av ApplicationVersion) ParsedValuesFile() (result map[string]any, err error) {
	if av.ValuesFileData != nil {
		if err = yaml.Unmarshal(av.ValuesFileData, &result); err != nil {
			err = fmt.Errorf("cannot parse ApplicationVersion values file: %w", err)
		}
	}
	return
}

func (av ApplicationVersion) ParsedTemplateFile() (result map[string]any, err error) {
	if av.TemplateFileData != nil {
		if err = yaml.Unmarshal(av.TemplateFileData, &result); err != nil {
			err = fmt.Errorf("cannot parse ApplicationVersion values template: %w", err)
		}
	}
	return
}

func (av ApplicationVersion) ParsedComposeFile() (result map[string]any, err error) {
	if av.ComposeFileData != nil {
		if err = yaml.Unmarshal(av.ComposeFileData, &result); err != nil {
			err = fmt.Errorf("cannot parse ApplicationVersion compose file: %w", err)
		}
	}
	return
}

func (av ApplicationVersion) Validate(deplType DeploymentType) error {
	if deplType == DeploymentTypeDocker {
		if av.ComposeFileData == nil {
			return errors.New("missing compose file")
		} else if av.ChartType != nil || av.ChartName != nil || av.ChartUrl != nil || av.ChartVersion != nil ||
			av.ValuesFileData != nil || av.TemplateFileData != nil {
			return errors.New("unexpected kubernetes specifics in docker application")
		}
	} else if deplType == DepolymentTypeKubernetes {
		if av.ChartType == nil || *av.ChartType == "" ||
			av.ChartUrl == nil || *av.ChartUrl == "" ||
			av.ChartVersion == nil || *av.ChartVersion == "" {
			return errors.New("not all of chart type, url and version are given")
		} else if *av.ChartType == HelmChartTypeRepository && (av.ChartName == nil || *av.ChartName == "") {
			return errors.New("missing chart name")
		} else if av.ComposeFileData != nil {
			return errors.New("unexpected docker file in kubernetes application")
		}
	}
	return nil
}

type Deployment struct {
	Base
	DeploymentTargetId   string  `db:"deployment_target_id" json:"deploymentTargetId"`
	ApplicationVersionId string  `db:"application_version_id" json:"applicationVersionId"`
	ReleaseName          *string `db:"release_name" json:"releaseName"`
	ValuesYaml           []byte  `db:"values_yaml" json:"valuesYaml"`
}

func (d Deployment) ParsedValuesFile() (result map[string]any, err error) {
	if d.ValuesYaml != nil {
		if err = yaml.Unmarshal(d.ValuesYaml, &result); err != nil {
			err = fmt.Errorf("cannot parse Deployment values file: %w", err)
		}
	}
	return
}

type DeploymentWithData struct {
	Deployment
	ApplicationId          string `db:"application_id" json:"applicationId"`
	ApplicationName        string `db:"application_name" json:"applicationName"`
	ApplicationVersionName string `db:"application_version_name" json:"applicationVersionName"`
}

type DeploymentStatus struct {
	Base
	DeploymentId string               `db:"deployment_id" json:"deploymentId"`
	Type         DeploymentStatusType `db:"type" json:"type"`
	Message      string               `db:"message" json:"message"`
}

type DeploymentTarget struct {
	Base
	Name                   string                  `db:"name" json:"name"`
	Type                   DeploymentType          `db:"type" json:"type"`
	Geolocation            *Geolocation            `db:"geolocation" json:"geolocation,omitempty"`
	AccessKeySalt          *[]byte                 `db:"access_key_salt" json:"-"`
	AccessKeyHash          *[]byte                 `db:"access_key_hash" json:"-"`
	CurrentStatus          *DeploymentTargetStatus `db:"current_status" json:"currentStatus,omitempty"`
	Namespace              *string                 `db:"namespace" json:"namespace"`
	OrganizationID         string                  `db:"organization_id" json:"-"`
	CreatedByUserAccountID string                  `db:"created_by_user_account_id" json:"-"`
	AgentVersionID         *string                 `db:"agent_version_id" json:"-"`
}

func (dt *DeploymentTarget) Validate() error {
	if dt.Type == DepolymentTypeKubernetes {
		if dt.Namespace == nil || *dt.Namespace == "" {
			return errors.New("DeploymentTarget with type \"kubernetes\" must not have empty namespace")
		}
	}
	return nil
}

type DeploymentTargetWithCreatedBy struct {
	DeploymentTarget
	CreatedBy        *UserAccountWithUserRole `db:"created_by" json:"createdBy"`
	LatestDeployment *DeploymentWithData      `db:"-" json:"latestDeployment,omitempty"`
	AgentVersion     *AgentVersion            `db:"agent_version" json:"agentVersion,omitempty"`
}

type DeploymentTargetStatus struct {
	// TODO unfortunately Base nested type doesn't work when ApplicationVersion is a nested row in an SQL query
	ID                 string    `db:"id" json:"id"`
	CreatedAt          time.Time `db:"created_at" json:"createdAt"`
	Message            string    `db:"message" json:"message"`
	DeploymentTargetId string    `db:"deployment_target_id" json:"-"`
}

type UserAccount struct {
	ID              string     `db:"id" json:"id"`
	CreatedAt       time.Time  `db:"created_at" json:"createdAt"`
	Email           string     `db:"email" json:"email"`
	EmailVerifiedAt *time.Time `db:"email_verified_at" json:"-"`
	PasswordHash    []byte     `db:"password_hash" json:"-"`
	PasswordSalt    []byte     `db:"password_salt" json:"-"`
	Name            string     `db:"name" json:"name,omitempty"`
	Password        string     `db:"-" json:"-"`
}

type UserAccountWithUserRole struct {
	// copy+pasted from UserAccount because pgx does not like embedded strucs
	ID              string     `db:"id" json:"id"`
	CreatedAt       time.Time  `db:"created_at" json:"createdAt"`
	Email           string     `db:"email" json:"email"`
	EmailVerifiedAt *time.Time `db:"email_verified_at" json:"-"`
	PasswordHash    []byte     `db:"password_hash" json:"-"`
	PasswordSalt    []byte     `db:"password_salt" json:"-"`
	Name            string     `db:"name" json:"name,omitempty"`
	UserRole        UserRole   `db:"user_role" json:"userRole"` // not copy+pasted
	Password        string     `db:"-" json:"-"`
}

type Organization struct {
	Base
	Name string `db:"name" json:"name"`
}

type OrganizationWithUserRole struct {
	Organization
	UserRole UserRole `db:"user_role"`
}

type UptimeMetric struct {
	Hour    time.Time `json:"hour"`
	Total   int       `json:"total"`
	Unknown int       `json:"unknown"`
}

type AgentVersion struct {
	Base
	Name                 string `db:"name" json:"name"`
	ManifestFileRevision string `db:"manifest_file_revision" json:"-"`
	ComposeFileRevision  string `db:"compose_file_revision" json:"-"`
}
