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

func DbAuthenticator() authn.Authenticator[AuthInfo, *DbAuthInfo] {
	return authn.AuthenticatorFunc[AuthInfo, *DbAuthInfo](func(ctx context.Context, a AuthInfo) (*DbAuthInfo, error) {
		var user *types.UserAccount
		var org *types.Organization
		var err error
		if a.CurrentOrgID() != nil && a.CurrentUserRole() != nil {
			if user, org, err = db.GetUserAccountAndOrg(
				ctx, a.CurrentUserID(), *a.CurrentOrgID(), *a.CurrentUserRole()); errors.Is(err, apierrors.ErrNotFound) {
				return nil, authn.ErrBadAuthentication
			}
		} else if user, err = db.GetUserAccountByID(ctx, a.CurrentUserID()); errors.Is(err, apierrors.ErrNotFound) {
			return nil, authn.ErrBadAuthentication
		}

		if err != nil {
			return nil, err
		}
		return &DbAuthInfo{
			AuthInfo: a,
			user:     user,
			org:      org,
		}, nil
	})
}
