package authz

import (
	"context"
	"errors"
	"slices"

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

	if name, err := name.Parse(nameStr); err != nil {
		return errors.New("name invalid")
	} else if org, err := db.GetOrganizationByID(ctx, *auth.CurrentOrgID()); err != nil {
		return err
	} else if org.ID.String() != name.OrgName {
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

	if name, err := name.Parse(nameStr); err != nil {
		return errors.New("name invalid")
	} else if org, err := db.GetOrganizationByID(ctx, *auth.CurrentOrgID()); err != nil {
		return err
	} else if org.ID.String() != name.OrgName {
		return ErrAccessDenied
	} else if action == ActionRead && *auth.CurrentUserRole() != types.UserRoleVendor {
		versions, err := db.GetLicensedArtifactVersions(ctx, name.OrgName, name.ArtifactName, auth.CurrentUserID())
		if err != nil {
			return err
		}
		if slices.ContainsFunc(versions, func(v types.ArtifactVersion) bool { return v.Name == reference }) {
			return ErrAccessDenied
		}
		// We might be serving a sub-manifest referenced by an index manifest, in which case the license might not be
		// for this manifest but for the index!
		// TODO: Check if there is a package version part for a version that would be licensed (similar to AuthorizeBlob)
	}
	return nil
}

// AuthorizeBlob implements ArtifactsAuthorizer.
func (a *authorizer) AuthorizeBlob(ctx context.Context, digest v1.Hash, action Action) error {
	auth := auth.ArtifactsAuthentication.Require(ctx)
	if action == ActionWrite && *auth.CurrentUserRole() != types.UserRoleVendor {
		return ErrAccessDenied
	}
	// TODO: Check if there is a package version in the org from the auth context that references this digest
	return nil
}
