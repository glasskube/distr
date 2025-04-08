package types

import (
	"time"

	"github.com/google/uuid"
)

type Application struct {
	ID             uuid.UUID            `db:"id" json:"id"`
	CreatedAt      time.Time            `db:"created_at" json:"createdAt"`
	OrganizationID uuid.UUID            `db:"organization_id" json:"-"`
	Name           string               `db:"name" json:"name"`
	Type           DeploymentType       `db:"type" json:"type"`
	Versions       []ApplicationVersion `db:"versions" json:"versions"`
}

type ApplicationWithIcon struct {
	Application
	Icon            []byte  `db:"icon" json:"icon"`
	IconFileName    *string `db:"icon_file_name" json:"iconFileName"`
	IconContentType *string `db:"icon_content_type" json:"iconContentType"`
}
