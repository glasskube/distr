package jwt

import (
	"context"
	"fmt"

	"github.com/glasskube/cloud/internal/authn"
	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

func Authenticator(ja *jwtauth.JWTAuth) authn.Authenticator[string, jwt.Token] {
	return authn.AuthenticatorFunc[string, jwt.Token](
		func(ctx context.Context, s string) (jwt.Token, error) {
			if token, err := jwtauth.VerifyToken(ja, s); err != nil {
				return nil, fmt.Errorf("%w: %w", authn.ErrBadAuthentication, err)
			} else {
				return token, nil
			}
		},
	)
}
