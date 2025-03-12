package manifest

import (
	"context"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/uuid"
)

type ManifestAuthChecker interface {
	Check(ctx context.Context, name string, userID, orgID uuid.UUID, action string) error
	CheckReference(ctx context.Context, name string, reference string, userID, orgID uuid.UUID, action string) error
	CheckBlob(ctx context.Context, digest v1.Hash, userID, orgID uuid.UUID, action string) error
}

type ManifestHandler interface {
	List(ctx context.Context, n int) ([]string, error)
	// ListTags
	//
	// Spec for implementation:
	// https://github.com/opencontainers/distribution-spec/blob/b505e9cc53ec499edbd9c1be32298388921bb705/detail.md#tags-paginated
	ListTags(ctx context.Context, name string, n int, last string) ([]string, error)
	ListDigests(ctx context.Context, name string) ([]v1.Hash, error)
	Get(ctx context.Context, name string, reference string) (*Manifest, error)
	Put(ctx context.Context, name string, reference string, manifest Manifest, blobs []v1.Hash) error
	Delete(ctx context.Context, name string, reference string) error
}
