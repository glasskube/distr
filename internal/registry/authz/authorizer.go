package authz

import (
	"context"
	"errors"

	"github.com/distr-sh/distr/internal/apierrors"
	"github.com/distr-sh/distr/internal/auth"
	"github.com/distr-sh/distr/internal/db"
	"github.com/distr-sh/distr/internal/registry/name"
	"github.com/distr-sh/distr/internal/types"
	"github.com/opencontainers/go-digest"
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
	AuthorizeBlob(ctx context.Context, digest digest.Digest, action Action) error
}

type authorizer struct{}

func NewAuthorizer() Authorizer {
	return &authorizer{}
}

// Authorize implements ArtifactsAuthorizer.
func (a *authorizer) Authorize(ctx context.Context, nameStr string, action Action) error {
	auth := auth.ArtifactsAuthentication.Require(ctx)

	if action == ActionWrite &&
		(auth.CurrentCustomerOrgID() != nil ||
			auth.CurrentUserRole() == nil ||
			*auth.CurrentUserRole() == types.UserRoleReadOnly) {
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

	if action == ActionWrite &&
		(auth.CurrentCustomerOrgID() != nil ||
			auth.CurrentUserRole() == nil ||
			*auth.CurrentUserRole() == types.UserRoleReadOnly) {
		return ErrAccessDenied
	}

	org := auth.CurrentOrg()
	if name, err := name.Parse(nameStr); err != nil {
		return err
	} else if org.Slug == nil || *org.Slug != name.OrgName {
		return ErrAccessDenied
	} else if action != ActionWrite && auth.CurrentCustomerOrgID() != nil {
		if org.HasFeature(types.FeatureLicensing) {
			err := db.CheckLicenseForArtifact(ctx,
				name.OrgName,
				name.ArtifactName,
				reference,
				*auth.CurrentCustomerOrgID(),
				*auth.CurrentOrgID(),
			)
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
func (a *authorizer) AuthorizeBlob(ctx context.Context, digest digest.Digest, action Action) error {
	auth := auth.ArtifactsAuthentication.Require(ctx)

	if action == ActionWrite &&
		(auth.CurrentCustomerOrgID() != nil ||
			auth.CurrentUserRole() == nil ||
			*auth.CurrentUserRole() == types.UserRoleReadOnly) {
		return ErrAccessDenied
	}

	if auth.CurrentCustomerOrgID() != nil && auth.CurrentOrg().HasFeature(types.FeatureLicensing) {
		err := db.CheckLicenseForArtifactBlob(ctx, digest.String(), *auth.CurrentCustomerOrgID(), *auth.CurrentOrgID())
		if errors.Is(err, apierrors.ErrForbidden) {
			return ErrAccessDenied
		} else if err != nil {
			return err
		}
	}

	return nil
}
