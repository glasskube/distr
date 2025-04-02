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
		authinfo.JWTAuthenticator(),
		authinfo.DbAuthenticator(),
	),
	authn.Chain4(
		token.NewExtractor(token.WithExtractorFuncs(token.FromHeader("AccessToken"))),
		authkey.Authenticator(),
		authinfo.AuthKeyAuthenticator(),
		authinfo.DbAuthenticator(),
	),
)

// AgentAuthentication supports Bearer JWT tokens
var AgentAuthentication = authn.New(
	authn.Chain3(
		token.NewExtractor(token.WithExtractorFuncs(token.FromHeader("Bearer"))),
		jwt.Authenticator(authjwt.JWTAuth),
		authinfo.JWTAuthenticator(),
		// TODO either a check for token audience or a new DbAuthenticator that verifies the given credentials against the DB
	),
)

// ArtifactsAuthentication supports Basic auth login for OCI clients, where the password should be a PAT.
// The given PAT is verified against the database, to make sure that the user still exists.
var ArtifactsAuthentication = authn.New(
	authn.Chain4(
		token.NewExtractor(
			token.WithExtractorFuncs(token.FromBasicAuth()),
			token.WithErrorHeaders(map[string]string{"WWW-Authenticate": "Basic realm=\"Distr\""}),
		),
		authkey.Authenticator(),
		authinfo.AuthKeyAuthenticator(),
		authinfo.DbAuthenticator(),
	),
)

func handleUnknownError(w http.ResponseWriter, r *http.Request, err error) {
	internalctx.GetLogger(r.Context()).Error("error authenticating request", zap.Error(err))
	sentry.GetHubFromContext(r.Context()).CaptureException(err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func init() {
	Authentication.SetUnknownErrorHandler(handleUnknownError)
	ArtifactsAuthentication.SetUnknownErrorHandler(handleUnknownError)
}
