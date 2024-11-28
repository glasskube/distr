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
	ApplicationId   string    `db:"application_id" json:"-"`
}

type DeploymentTarget struct {
	Base
	Name          string                  `db:"name" json:"name"`
	Type          DeploymentType          `db:"type" json:"type"`
	Geolocation   *Geolocation            `db:"geolocation" json:"geolocation,omitempty"`
	CurrentStatus *DeploymentTargetStatus `db:"current_status" json:"currentStatus,omitempty"`
}

type DeploymentTargetStatus struct {
	// TODO unfortunately Base nested type doesn't work when ApplicationVersion is a nested row in an SQL query
	ID                 string    `db:"id" json:"id"`
	CreatedAt          time.Time `db:"created_at" json:"createdAt"`
	Message            string    `db:"message" json:"message"`
	DeploymentTargetId string    `db:"deployment_target_id" json:"-"`
}
