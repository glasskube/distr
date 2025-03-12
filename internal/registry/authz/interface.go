package authz

import (
	"context"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/uuid"
)

type ArtifactsAuthorizer interface {
	Authorize(ctx context.Context, name string, userID, orgID uuid.UUID, action string) error
	AuthorizeReference(ctx context.Context, name string, reference string, userID, orgID uuid.UUID, action string) error
	AuthorizeBlob(ctx context.Context, digest v1.Hash, userID, orgID uuid.UUID, action string) error
}
