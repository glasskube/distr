package authz

import (
	"context"
	"errors"

	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/auth"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/registry/name"
	"github.com/glasskube/distr/internal/types"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type Action string

const (
	ActionRead  Action = "read"
	ActionWrite Action = "write"
	ActionStat  Action = "stat"
)

type Authorizer interface {
	Authorize(ctx context.Context, name string, action Action) error
	AuthorizeReference(ctx context.Context, name string, reference string, action Action) error
	AuthorizeBlob(ctx context.Context, digest v1.Hash, action Action) error
}

type authorizer struct{}

func NewAuthorizer() Authorizer {
	return &authorizer{}
}

// Authorize implements ArtifactsAuthorizer.
func (a *authorizer) Authorize(ctx context.Context, nameStr string, action Action) error {
	auth := auth.ArtifactsAuthentication.Require(ctx)
	if action == ActionWrite && *auth.CurrentUserRole() != types.UserRoleVendor {
		return ErrAccessDenied
	}

	org := auth.CurrentOrg()
	if name, err := name.Parse(nameStr); err != nil {
		return err
	} else if org.Slug == nil || *org.Slug != name.OrgName {
		return ErrAccessDenied
	}
	return nil
}

// AuthorizeReference implements ArtifactsAuthorizer.
func (a *authorizer) AuthorizeReference(ctx context.Context, nameStr string, reference string, action Action) error {
	auth := auth.ArtifactsAuthentication.Require(ctx)
	if action == ActionWrite && *auth.CurrentUserRole() != types.UserRoleVendor {
		return ErrAccessDenied
	}

	org := auth.CurrentOrg()
	if name, err := name.Parse(nameStr); err != nil {
		return err
	} else if org.Slug == nil || *org.Slug != name.OrgName {
		return ErrAccessDenied
	} else if action != ActionWrite && *auth.CurrentUserRole() != types.UserRoleVendor {
		if org.HasFeature(types.FeatureLicensing) {
			err := db.CheckLicenseForArtifact(ctx, name.OrgName, name.ArtifactName, reference, auth.CurrentUserID())
			if errors.Is(err, apierrors.ErrForbidden) {
				return ErrAccessDenied
			} else if err != nil {
				return err
			}
		}
	}
	return nil
}

// AuthorizeBlob implements ArtifactsAuthorizer.
func (a *authorizer) AuthorizeBlob(ctx context.Context, digest v1.Hash, action Action) error {
	auth := auth.ArtifactsAuthentication.Require(ctx)

	if *auth.CurrentUserRole() != types.UserRoleVendor {
		if action == ActionWrite {
			return ErrAccessDenied
		} else if auth.CurrentOrg().HasFeature(types.FeatureLicensing) {
			err := db.CheckLicenseForArtifactBlob(ctx, digest.String(), auth.CurrentUserID())
			if errors.Is(err, apierrors.ErrForbidden) {
				return ErrAccessDenied
			} else if err != nil {
				return err
			}
		}
	}
	return nil
}
