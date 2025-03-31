package authinfo

import (
	"context"
	"errors"
	"fmt"

	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/authkey"
	"github.com/glasskube/distr/internal/authn"
	"github.com/glasskube/distr/internal/db"
)

func FromAuthKey(ctx context.Context, token authkey.Key) (AuthInfo, error) {
	if at, err := db.GetAccessTokenByKeyUpdatingLastUsed(ctx, token); err != nil {
		if errors.Is(err, apierrors.ErrNotFound) {
			err = fmt.Errorf("%w: %w", authn.ErrBadAuthentication, err)
		}
		return nil, err
	} else if orgs, err := db.GetOrganizationsForUser(ctx, at.UserAccount.ID); err != nil {
		return nil, err
	} else if len(orgs) != 1 {
		return nil, errors.New("user must have exacly one organization")
	} else {
		org := orgs[0]
		// TODO get user + db and use DbAuthInfo instead
		return &SimpleAuthInfo{
			userID:         at.UserAccount.ID,
			userEmail:      at.UserAccount.Email,
			emailVerified:  at.UserAccount.EmailVerifiedAt != nil,
			organizationID: &org.ID,
			userRole:       &org.UserRole,
			rawToken:       token,
		}, nil
	}
}

func AuthKeyAuthenticator() authn.Authenticator[authkey.Key, AuthInfo] {
	return authn.AuthenticatorFunc[authkey.Key, AuthInfo](
		func(ctx context.Context, key authkey.Key) (AuthInfo, error) {
			return FromAuthKey(ctx, key)
		},
	)
}
