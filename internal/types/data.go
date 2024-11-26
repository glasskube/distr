package types

type Application struct {
	Base
	Name string         `db:"name" json:"name"`
	Type DeploymentType `db:"type" json:"type"`
}

type DeploymentTarget struct {
	Base
	Name        string         `db:"name" json:"name"`
	Type        DeploymentType `db:"type" json:"type"`
	Geolocation *Geolocation   `db:"geolocation" json:"geolocation,omitempty"`
}
