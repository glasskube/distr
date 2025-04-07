package auth

import (
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/internal/authjwt"
	"github.com/glasskube/distr/internal/authn"
	"github.com/glasskube/distr/internal/authn/authinfo"
	"github.com/glasskube/distr/internal/authn/authkey"
	"github.com/glasskube/distr/internal/authn/jwt"
	"github.com/glasskube/distr/internal/authn/token"
	internalctx "github.com/glasskube/distr/internal/context"
	"go.uber.org/zap"
)

// Authentication supports Bearer (classic JWT) and AccessToken (PAT) headers and uses authinfo.DbAuthenticator
// to verify the token against the database, thereby ensuring the user exists in the database.
var Authentication = authn.New(
	authn.Chain4(
		token.NewExtractor(token.WithExtractorFuncs(token.FromHeader("Bearer"))),
		jwt.Authenticator(authjwt.JWTAuth),
		authinfo.UserJWTAuthenticator(),
		authinfo.DbAuthenticator(),
	),
	authn.Chain4(
		token.NewExtractor(token.WithExtractorFuncs(token.FromHeader("AccessToken"))),
		authkey.Authenticator(),
		authinfo.AuthKeyAuthenticator(),
		authinfo.DbAuthenticator(),
	),
)

// AgentAuthentication supports only Bearer JWT tokens
var AgentAuthentication = authn.New(
	authn.Chain3(
		token.NewExtractor(token.WithExtractorFuncs(token.FromHeader("Bearer"))),
		jwt.Authenticator(authjwt.JWTAuth),
		authinfo.AgentJWTAuthenticator(),
		// for agents, db check is done in the agent auth middleware, therefore no DbAuthenticator here
	),
)

// ArtifactsAuthentication supports Basic auth login for OCI clients, where the password should be a PAT.
// The given PAT is verified against the database, to make sure that the user still exists.
var ArtifactsAuthentication = authn.New(
	authn.Chain(
		token.NewExtractor(
			token.WithExtractorFuncs(token.FromBasicAuth()),
			token.WithErrorHeaders(http.Header{"WWW-Authenticate": []string{"Basic realm=\"Distr\""}}),
		),
		authn.Alternative(
			// Auhtenticate UserAccount with PAT
			authn.Chain3(
				authkey.Authenticator(),
				authinfo.AuthKeyAuthenticator(),
				authinfo.DbAuthenticator(),
			),
			// Authenticate with Agent JWT
			authn.Chain3(
				jwt.Authenticator(authjwt.JWTAuth),
				authinfo.AgentJWTAuthenticator(),
				authinfo.AgentDbAuthenticator(),
			),
		),
	),
)

func handleUnknownError(w http.ResponseWriter, r *http.Request, err error) {
	internalctx.GetLogger(r.Context()).Error("error authenticating request", zap.Error(err))
	sentry.GetHubFromContext(r.Context()).CaptureException(err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func init() {
	Authentication.SetUnknownErrorHandler(handleUnknownError)
	AgentAuthentication.SetUnknownErrorHandler(handleUnknownError)
	ArtifactsAuthentication.SetUnknownErrorHandler(handleUnknownError)
}
