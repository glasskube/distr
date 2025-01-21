package token

import (
	"context"
	"net/http"
	"strings"

	"github.com/glasskube/cloud/internal/authn"
)

type TokenExtractorFunc func(r *http.Request) string

type TokenExtractor []TokenExtractorFunc

// Authenticate implements Provider.
func (fns TokenExtractor) Authenticate(ctx context.Context, r *http.Request) (string, error) {
	for _, fn := range fns {
		if token := fn(r); token != "" {
			return token, nil
		}
	}
	return "", authn.ErrNoAuthentication
}

var _ authn.RequestAuthenticator[string] = TokenExtractor{}

func NewTokenExtractor(tokenFuncs ...TokenExtractorFunc) TokenExtractor {
	return TokenExtractor(tokenFuncs)
}

func TokenFromQuery(param string) TokenExtractorFunc {
	return func(r *http.Request) string { return r.URL.Query().Get(param) }
}

func TokenFromHeader(authenticationScheme string) TokenExtractorFunc {
	prefix := strings.ToUpper(authenticationScheme + " ")
	return func(r *http.Request) string {
		authorization := r.Header.Get("Authorization")
		if len(authorization) > len(prefix) && strings.ToUpper(authorization[0:len(prefix)]) == prefix {
			return authorization[len(prefix):]
		}
		return ""
	}
}
