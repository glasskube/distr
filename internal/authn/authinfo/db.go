package authinfo

import (
	"context"
	"errors"

	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/authn"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/types"
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

func DbAuthenticator() authn.Authenticator[AuthInfo, AuthInfo] {
	return authn.AuthenticatorFunc[AuthInfo, AuthInfo](func(ctx context.Context, a AuthInfo) (AuthInfo, error) {
		var user *types.UserAccount
		var org *types.Organization
		var err error
		if a.CurrentOrgID() == nil {
			if user, err = db.GetUserAccountByID(ctx, a.CurrentUserID()); errors.Is(err, apierrors.ErrNotFound) {
				return nil, authn.ErrNoAuthentication
			}
		} else {
			var userWithRole *types.UserAccountWithUserRole
			if userWithRole, org, err = db.GetUserAccountAndOrgWithRole(
				ctx, a.CurrentUserID(), *a.CurrentOrgID()); errors.Is(err, apierrors.ErrNotFound) {
				return nil, authn.ErrBadAuthentication
			} else if err != nil {
				return nil, err
			} else if a.CurrentUserRole() != nil && userWithRole.UserRole != *a.CurrentUserRole() {
				return nil, authn.ErrBadAuthentication
			}
		}
		return &DbAuthInfo{
			AuthInfo: a,
			user:     user,
			org:      org,
		}, nil
	})
}
