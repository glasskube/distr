package types

import "time"

type Application struct {
	Base
	Name     string               `db:"name" json:"name"`
	Type     DeploymentType       `db:"type" json:"type"`
	Versions []ApplicationVersion `db:"versions" json:"versions"`
}

type ApplicationVersion struct {
	// TODO unfortunately Base nested type doesn't work when ApplicationVersion is a nested row in an SQL query
	ID              string    `db:"id" json:"id"`
	CreatedAt       time.Time `db:"created_at" json:"createdAt"`
	Name            string    `db:"name" json:"name"`
	ComposeFileData *[]byte   `db:"compose_file_data" json:"-"`
	ApplicationId   string    `db:"application_id" json:"applicationId"`
}

type Deployment struct {
	Base
	DeploymentTargetId   string `db:"deployment_target_id" json:"deploymentTargetId"`
	ApplicationVersionId string `db:"application_version_id" json:"applicationVersionId"`
}

type DeploymentWithData struct {
	Deployment
	ApplicationId          string `db:"application_id" json:"applicationId"`
	ApplicationName        string `db:"application_name" json:"applicationName"`
	ApplicationVersionName string `db:"application_version_name" json:"applicationVersionName"`
}

type DeploymentTarget struct {
	Base
	Name          string                  `db:"name" json:"name"`
	Type          DeploymentType          `db:"type" json:"type"`
	Geolocation   *Geolocation            `db:"geolocation" json:"geolocation,omitempty"`
	AccessKeySalt *[]byte                 `db:"access_key_salt"`
	AccessKeyHash *[]byte                 `db:"access_key_hash"`
	CurrentStatus *DeploymentTargetStatus `db:"current_status" json:"currentStatus,omitempty"`
}

type DeploymentTargetStatus struct {
	// TODO unfortunately Base nested type doesn't work when ApplicationVersion is a nested row in an SQL query
	ID                 string    `db:"id" json:"id"`
	CreatedAt          time.Time `db:"created_at" json:"createdAt"`
	Message            string    `db:"message" json:"message"`
	DeploymentTargetId string    `db:"deployment_target_id" json:"-"`
}

type UserAccount struct {
	Base
	Email        string `db:"email"`
	PasswordHash []byte `db:"password_hash"`
	PasswordSalt []byte `db:"password_salt"`
	Password     string `db:"-"`
	Name         string `db:"name"`
}

type Organization struct {
	Base
	Name string `db:"name" json:"name"`
}
