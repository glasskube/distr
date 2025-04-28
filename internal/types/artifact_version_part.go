package types

import (
	"github.com/google/uuid"
)

type ArtifactVersionPart struct {
	ArtifactVersionID  uuid.UUID `db:"artifact_version_id" json:"-"`
	ArtifactBlobDigest Digest    `db:"artifact_blob_digest" json:"-"`
	ArtifactBlobSize   int64     `db:"artifact_blob_size"`
}
