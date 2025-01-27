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

var Authentication = authn.New(
	authn.Chain(
		authn.Chain(
			token.NewExtractor(token.FromHeader("Bearer")),
			jwt.Authenticator(authjwt.JWTAuth),
		),
		authinfo.JWTAuthenticator(),
	),
	authn.Chain(
		authn.Chain(
			token.NewExtractor(token.FromHeader("AccessToken")),
			authkey.Authenticator(),
		),
		authinfo.AuthKeyAuthenticator(),
	),
)

func init() {
	Authentication.SetUnknownErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
		internalctx.GetLogger(r.Context()).Error("error authenticating request", zap.Error(err))
		sentry.GetHubFromContext(r.Context()).CaptureException(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	})
}
