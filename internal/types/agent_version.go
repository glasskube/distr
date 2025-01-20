package types

import "time"

type AgentVersion struct {
	ID                   string    `db:"id" json:"id"`
	CreatedAt            time.Time `db:"created_at" json:"createdAt"`
	Name                 string    `db:"name" json:"name"`
	ManifestFileRevision string    `db:"manifest_file_revision" json:"-"`
	ComposeFileRevision  string    `db:"compose_file_revision" json:"-"`
}
