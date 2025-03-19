package auth

import (
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/internal/authjwt"
	"github.com/glasskube/distr/internal/authn"
	"github.com/glasskube/distr/internal/authn/agent"
	"github.com/glasskube/distr/internal/authn/authinfo"
	"github.com/glasskube/distr/internal/authn/authkey"
	"github.com/glasskube/distr/internal/authn/basic"
	"github.com/glasskube/distr/internal/authn/jwt"
	"github.com/glasskube/distr/internal/authn/token"
	internalctx "github.com/glasskube/distr/internal/context"
	"go.uber.org/zap"
)

var Authentication = authn.New(
	authn.Chain3(
		token.NewExtractor(token.WithExtractorFuncs(token.FromHeader("Bearer"))),
		jwt.Authenticator(authjwt.JWTAuth),
		authinfo.JWTAuthenticator(),
	),
	authn.Chain3(
		token.NewExtractor(token.WithExtractorFuncs(token.FromHeader("AccessToken"))),
		authkey.Authenticator(),
		authinfo.AuthKeyAuthenticator(),
	),
)

var ArtifactsAuthentication = authn.New(
	authn.Alternative(
		authn.Chain3(
			token.NewExtractor(token.WithExtractorFuncs(token.FromHeader("Bearer"))),
			jwt.Authenticator(authjwt.JWTAuth),
			authinfo.JWTAuthenticator(),
		),
		authn.Chain(
			basic.Authenticator(),
			authn.Alternative(
				authn.Chain3(
					basic.Password(),
					authkey.Authenticator(),
					authinfo.AuthKeyAuthenticator(),
				),
				agent.Authenticator(),
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
	ArtifactsAuthentication.SetUnknownErrorHandler(handleUnknownError)
}
