package types

import (
	"github.com/google/uuid"
	"time"
)

type ArtifactVersionPart struct {
	ID                     uuid.UUID  `db:"id" json:"id"`
	CreatedAt              time.Time  `db:"created_at" json:"createdAt"`
	CreatedByUserAccountID *uuid.UUID `db:"created_by_useraccount_id" json:"-"`
	HashMD5                string     `db:"hash_md5" json:"hashMD5"`
	HashSha1               string     `db:"hash_sha1" json:"hashSha1"`
	HashSha256             string     `db:"hash_sha256" json:"hashSha256"`
	HashSha512             string     `db:"hash_sha512" json:"hashSha512"`

	ArtifactVersionID uuid.UUID `db:"artifact_version_id" json:"artifactVersionId"`
	ArtifactBlobID    uuid.UUID `db:"artifact_blob_id" json:"artifactBlobId"`
}
