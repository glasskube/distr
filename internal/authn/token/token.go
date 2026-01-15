package token

import (
	"context"
	"net/http"
	"strings"

	"github.com/distr-sh/distr/internal/authn"
)

type TokenExtractorFunc func(r *http.Request) string

type TokenExtractor struct {
	fns     []TokenExtractorFunc
	headers http.Header
}

// Authenticate implements Provider.
func (extractor *TokenExtractor) Authenticate(ctx context.Context, r *http.Request) (string, error) {
	for _, fn := range extractor.fns {
		if token := fn(r); token != "" {
			return token, nil
		}
	}
	return "", authn.NewHttpHeaderError(authn.ErrNoAuthentication, extractor.headers)
}

var _ authn.RequestAuthenticator[string] = &TokenExtractor{}

type ExtractorOption func(*TokenExtractor)

func WithExtractorFuncs(fns ...TokenExtractorFunc) ExtractorOption {
	return func(te *TokenExtractor) {
		te.fns = append(te.fns, fns...)
	}
}

func WithErrorHeaders(headers http.Header) ExtractorOption {
	return func(te *TokenExtractor) {
		te.headers = headers
	}
}

func NewExtractor(opts ...ExtractorOption) *TokenExtractor {
	extractor := TokenExtractor{}
	for _, opt := range opts {
		opt(&extractor)
	}
	return &extractor
}

func FromQuery(param string) TokenExtractorFunc {
	return func(r *http.Request) string { return r.URL.Query().Get(param) }
}

func FromHeader(authenticationScheme string) TokenExtractorFunc {
	prefix := strings.ToUpper(authenticationScheme + " ")
	return func(r *http.Request) string {
		authorization := r.Header.Get("Authorization")
		if len(authorization) > len(prefix) && strings.ToUpper(authorization[0:len(prefix)]) == prefix {
			return authorization[len(prefix):]
		}
		return ""
	}
}

func FromBasicAuth() TokenExtractorFunc {
	return func(r *http.Request) string {
		if _, token, ok := r.BasicAuth(); ok {
			return token
		} else {
			return ""
		}
	}
}
