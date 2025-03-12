package types

import (
	"time"

	"github.com/google/uuid"
)

type ArtifactVersion struct {
	ID                     uuid.UUID  `db:"id" json:"id"`
	CreatedAt              time.Time  `db:"created_at" json:"createdAt"`
	CreatedByUserAccountID *uuid.UUID `db:"created_by_useraccount_id" json:"-"`
	UpdatedAt              *time.Time `db:"updated_at" json:"updatedAt"`
	UpdatedByUserAccountID *uuid.UUID `db:"updated_by_useraccount_id" json:"-"`
	Name                   string     `db:"name" json:"name"`
	ManifestBlobDigest     string     `db:"manifest_blob_digest" json:"manifestBlobDigest"`
	ManifestContentType    string     `db:"manifest_content_type" json:"manifestContentType"`
	ArtifactID             uuid.UUID  `db:"artifact_id" json:"artifactId"`
}
