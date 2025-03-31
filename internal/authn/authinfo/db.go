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
	user *types.UserAccountWithUserRole
	org  *types.Organization
}

func (a DbAuthInfo) CurrentUser() *types.UserAccountWithUserRole {
	return a.user
}

func (a DbAuthInfo) CurrentOrg() *types.Organization {
	return a.org
}

func DbAuthenticator() authn.Authenticator[AuthInfo, AuthInfo] {
	return authn.AuthenticatorFunc[AuthInfo, AuthInfo](func(ctx context.Context, a AuthInfo) (AuthInfo, error) {
		// TODO also get org (possibly in the same db query?)
		// TODO also check: user still in that org with that role?
		if a.CurrentOrgID() == nil {
			// TODO check: some tokens (e.g. invite) don't have the org ID?
			return a, nil
		}
		if user, org, err := db.GetUserAccountAndOrgWithRole(ctx, a.CurrentUserID(), *a.CurrentOrgID()); errors.Is(err, apierrors.ErrNotFound) {
			return nil, authn.ErrBadAuthentication
		} else if err != nil {
			return nil, err
		} else if user.UserRole != *a.CurrentUserRole() {
			return nil, authn.ErrBadAuthentication // TODO ??
		} else {
			return &DbAuthInfo{
				AuthInfo: a,
				user:     user,
				org:      org,
			}, nil
		}
	})
}
