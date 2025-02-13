package types

import "github.com/google/uuid"

type Application struct {
	Base
	OrganizationID uuid.UUID            `db:"organization_id" json:"-"`
	Name           string               `db:"name" json:"name"`
	Type           DeploymentType       `db:"type" json:"type"`
	Versions       []ApplicationVersion `db:"versions" json:"versions"`
}
