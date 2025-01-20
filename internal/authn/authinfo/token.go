package authinfo

import (
	"context"
	"errors"
	"fmt"

	"github.com/glasskube/cloud/internal/apierrors"
	"github.com/glasskube/cloud/internal/authkey"
	"github.com/glasskube/cloud/internal/authn"
	"github.com/glasskube/cloud/internal/db"
)

func FromAuthKey(ctx context.Context, token authkey.Key) (AuthInfo, error) {
	if at, err := db.GetAccessTokenByKey(ctx, token); err != nil {
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
		return &SimpleAuthInfo{
			userID:         at.UserAccount.ID,
			emailVerified:  at.UserAccount.EmailVerifiedAt != nil,
			organizationID: org.ID,
			userRole:       &org.UserRole,
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
