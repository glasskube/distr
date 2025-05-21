package authinfo

import (
	"context"
	"errors"

	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/authn"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/util"
)

type DbAuthInfo struct {
	AuthInfo
	user *types.UserAccount
	org  *types.Organization
}

func (a DbAuthInfo) CurrentUser() *types.UserAccount {
	return a.user
}

func (a DbAuthInfo) CurrentOrg() *types.Organization {
	return a.org
}

func DbAuthenticator() authn.Authenticator[AuthInfo, *DbAuthInfo] {
	return authn.AuthenticatorFunc[AuthInfo, *DbAuthInfo](func(ctx context.Context, a AuthInfo) (*DbAuthInfo, error) {
		var user *types.UserAccount
		var org *types.Organization
		if a.CurrentOrgID() != nil && a.CurrentUserRole() != nil {
			if u, o, err := db.GetUserAccountAndOrg(
				ctx, a.CurrentUserID(), *a.CurrentOrgID(), a.CurrentUserRole()); errors.Is(err, apierrors.ErrNotFound) {
				return nil, authn.ErrBadAuthentication
			} else if err != nil {
				return nil, err
			} else {
				user = util.PtrTo(u.AsUserAccount())
				org = o
			}
		} else {
			// some special tokens like password reset don't have an organization ID
			if u, err := db.GetUserAccountByID(ctx, a.CurrentUserID()); errors.Is(err, apierrors.ErrNotFound) {
				return nil, authn.ErrBadAuthentication
			} else if err != nil {
				return nil, err
			} else {
				user = u
			}
		}
		return &DbAuthInfo{
			AuthInfo: a,
			user:     user,
			org:      org,
		}, nil
	})
}

func AgentDbAuthenticator() authn.Authenticator[AgentAuthInfo, *DbAuthInfo] {
	fn := func(ctx context.Context, a AgentAuthInfo) (*DbAuthInfo, error) {
		userWithRole, org, err := db.GetUserAccountAndOrgForDeploymentTarget(ctx, a.CurrentDeploymentTargetID())
		if errors.Is(err, apierrors.ErrNotFound) {
			return nil, authn.ErrBadAuthentication
		} else if err != nil {
			return nil, err
		}
		return &DbAuthInfo{
			AuthInfo: &SimpleAuthInfo{
				organizationID: &org.ID,
				userID:         userWithRole.ID,
				userEmail:      userWithRole.Email,
				emailVerified:  userWithRole.EmailVerifiedAt != nil,
				userRole:       util.PtrTo(userWithRole.UserRole),
				rawToken:       a.Token(),
			},
			user: util.PtrTo(userWithRole.AsUserAccount()),
			org:  org,
		}, nil
	}
	return authn.AuthenticatorFunc[AgentAuthInfo, *DbAuthInfo](fn)
}
