package types

import "time"

type Application struct {
	ID        string               `db:"id" json:"id"`
	CreatedAt time.Time            `db:"created_at" json:"createdAt"`
	Name      string               `db:"name" json:"name"`
	Type      DeploymentType       `db:"type" json:"type"`
	Versions  []ApplicationVersion `db:"-" json:"versions"`
}

type ApplicationVersion struct {
	ID              string    `db:"id" json:"id"`
	CreatedAt       time.Time `db:"created_at" json:"createdAt"`
	Name            string    `db:"name" json:"name"`
	ComposeFileData *[]byte   `db:"compose_file_data" json:"composeFileData"`
	ApplicationId   string    `db:"application_id" json:"-"`
}
