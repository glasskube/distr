package authkey

import (
	"context"
	"fmt"

	"github.com/glasskube/distr/internal/authkey"
	"github.com/glasskube/distr/internal/authn"
)

func Authenticator() authn.Authenticator[string, authkey.Key] {
	return authn.AuthenticatorFunc[string, authkey.Key](
		func(ctx context.Context, token string) (authkey.Key, error) {
			if key, err := authkey.Parse(token); err != nil {
				return authkey.Key{}, fmt.Errorf("%w: %w", authn.ErrBadAuthentication, err)
			} else {
				return key, nil
			}
		},
	)
}
