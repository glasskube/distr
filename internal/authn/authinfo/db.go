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
		if a.CurrentOrgID() != nil {
			if user, org, err = db.GetUserAccountAndOrg(
				ctx, a.CurrentUserID(), *a.CurrentOrgID(), *a.CurrentUserRole()); errors.Is(err, apierrors.ErrNotFound) {
				return nil, authn.ErrBadAuthentication
			}
		} else if user, err = db.GetUserAccountByID(ctx, a.CurrentUserID()); errors.Is(err, apierrors.ErrNotFound) {
			return nil, authn.ErrNoAuthentication
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
