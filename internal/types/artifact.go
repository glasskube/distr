package types

import (
	"time"

	"github.com/google/uuid"
)

type Artifact struct {
	ID             uuid.UUID `db:"id" json:"id"`
	CreatedAt      time.Time `db:"created_at" json:"createdAt"`
	OrganizationID uuid.UUID `db:"organization_id" json:"-"`
	Name           string    `db:"name" json:"name"`
}

type ArtifactVersionTag struct {
	ID   uuid.UUID `db:"id" json:"id"`
	Name string    `db:"name" json:"name"`
}

type Downloads struct {
	DownloadsTotal    int         `db:"downloads_total" json:"downloadsTotal"`
	DownloadedByCount int         `db:"downloaded_by_count" json:"downloadedByCount"`
	DownloadedByUsers []uuid.UUID `db:"downloaded_by_users" json:"downloadedByUsers,omitempty"`
}

type TaggedArtifactVersion struct {
	ID        uuid.UUID            `db:"id" json:"id"`
	CreatedAt time.Time            `db:"created_at" json:"createdAt"`
	Digest    string               `db:"manifest_blob_digest" json:"digest"`
	Tags      []ArtifactVersionTag `db:"tags" json:"tags"`
	Size      int64                `db:"size" json:"size"`
	Downloads Downloads            `db:"downloads" json:"downloads"`
}

type ArtifactWithDownloads struct {
	Artifact
	OrganizationSlug string    `db:"organization_slug" json:"-"`
	Downloads        Downloads `db:"downloads" json:"downloads"`
}

type ArtifactWithTaggedVersion struct {
	ArtifactWithDownloads
	Versions []TaggedArtifactVersion `db:"versions" json:"versions,omitempty"`
}
