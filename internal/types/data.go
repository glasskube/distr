package types

type Application struct {
	Base
	Name      string               `db:"name" json:"name"`
	Type      DeploymentType       `db:"type" json:"type"`
	Versions  []ApplicationVersion `db:"versions" json:"versions"`
}

type ApplicationVersion struct {
	Base
	Name            string    `db:"name" json:"name"`
	ComposeFileData *[]byte   `db:"compose_file_data" json:"-"`
	ApplicationId   string    `db:"application_id" json:"-"`
}

type DeploymentTarget struct {
	Base
	Name        string         `db:"name" json:"name"`
	Type        DeploymentType `db:"type" json:"type"`
	Geolocation *Geolocation   `db:"geolocation" json:"geolocation,omitempty"`
}
