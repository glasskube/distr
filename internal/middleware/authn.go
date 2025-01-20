package middleware

import (
	"github.com/glasskube/cloud/internal/authjwt"
	"github.com/glasskube/cloud/internal/authn"
	"github.com/glasskube/cloud/internal/authn/authinfo"
	"github.com/glasskube/cloud/internal/authn/authkey"
	"github.com/glasskube/cloud/internal/authn/jwt"
	"github.com/glasskube/cloud/internal/authn/token"
)

var Authn = authn.New(
	authn.Chain(
		authn.Chain(
			token.NewTokenExtractor(
				token.TokenFromHeader("Bearer"),
				token.TokenFromQuery("jwt"),
			),
			jwt.Authenticator(authjwt.JWTAuth),
		),
		authinfo.JWTAuthenticator(),
	),
	authn.Chain(
		authn.Chain(
			token.NewTokenExtractor(
				token.TokenFromHeader("AccessToken"),
				token.TokenFromQuery("token"),
			),
			authkey.Authenticator(),
		),
		authinfo.AuthKeyAuthenticator(),
	),
)
