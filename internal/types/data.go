package types

import "time"

type Application struct {
	Base
	OrganizationID string               `db:"organization_id" json:"-"`
	Name           string               `db:"name" json:"name"`
	Type           DeploymentType       `db:"type" json:"type"`
	Versions       []ApplicationVersion `db:"versions" json:"versions"`
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
	Name           string                  `db:"name" json:"name"`
	Type           DeploymentType          `db:"type" json:"type"`
	Geolocation    *Geolocation            `db:"geolocation" json:"geolocation,omitempty"`
	AccessKeySalt  *[]byte                 `db:"access_key_salt" json:"-"`
	AccessKeyHash  *[]byte                 `db:"access_key_hash" json:"-"`
	CurrentStatus  *DeploymentTargetStatus `db:"current_status" json:"currentStatus,omitempty"`
	OrganizationID string                  `db:"organization_id" json:"-"`
}

type DeploymentTargetWithCreatedBy struct {
	DeploymentTarget
	CreatedBy *UserAccountWithUserRole `db:"created_by" json:"createdBy"`
}

type DeploymentTargetStatus struct {
	// TODO unfortunately Base nested type doesn't work when ApplicationVersion is a nested row in an SQL query
	ID                 string    `db:"id" json:"id"`
	CreatedAt          time.Time `db:"created_at" json:"createdAt"`
	Message            string    `db:"message" json:"message"`
	DeploymentTargetId string    `db:"deployment_target_id" json:"-"`
}

type UserAccount struct {
	ID           string    `db:"id" json:"id"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
	Email        string    `db:"email" json:"email"`
	PasswordHash []byte    `db:"password_hash" json:"-"`
	PasswordSalt []byte    `db:"password_salt" json:"-"`
	Name         string    `db:"name" json:"name,omitempty"`
	Password     string    `db:"-" json:"-"`
}

type UserAccountWithUserRole struct {
	// copy+pasted from UserAccount because pgx does not like embedded strucs
	ID           string    `db:"id" json:"id"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
	Email        string    `db:"email" json:"email"`
	PasswordHash []byte    `db:"password_hash" json:"-"`
	PasswordSalt []byte    `db:"password_salt" json:"-"`
	Name         string    `db:"name" json:"name,omitempty"`
	UserRole     UserRole  `db:"user_role" json:"userRole"` // not copy+pasted
	Password     string    `db:"-" json:"-"`
}

type Organization struct {
	Base
	Name string `db:"name" json:"name"`
}

type OrganizationWithUserRole struct {
	Organization
	UserRole UserRole `db:"user_role"`
}
