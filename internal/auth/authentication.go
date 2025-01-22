package auth

import (
	"github.com/glasskube/cloud/internal/authjwt"
	"github.com/glasskube/cloud/internal/authn"
	"github.com/glasskube/cloud/internal/authn/authinfo"
	"github.com/glasskube/cloud/internal/authn/authkey"
	"github.com/glasskube/cloud/internal/authn/jwt"
	"github.com/glasskube/cloud/internal/authn/token"
)

var Authentication = authn.New(
	authn.Chain(
		authn.Chain(
			token.NewTokenExtractor(token.TokenFromHeader("Bearer")),
			jwt.Authenticator(authjwt.JWTAuth),
		),
		authinfo.JWTAuthenticator(),
	),
	authn.Chain(
		authn.Chain(
			token.NewTokenExtractor(token.TokenFromHeader("AccessToken")),
			authkey.Authenticator(),
		),
		authinfo.AuthKeyAuthenticator(),
	),
)
