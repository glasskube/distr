package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/distr-sh/distr/internal/authkey"
	"github.com/distr-sh/distr/internal/util"
	"go.uber.org/zap"
)

var defaultBaseUrl = util.Require(url.Parse("https://app.distr.sh/"))

type authTokenKey struct{}

func AuthTokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(authTokenKey{}).(string)
	return token, ok
}

func WithAuthToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, authTokenKey{}, token)
}

type ConfigOption func(*Config)

func WithToken(token authkey.Key) ConfigOption {
	return func(c *Config) {
		c.token = token
	}
}

func WithBaseURL(url *url.URL) ConfigOption {
	return func(c *Config) {
		c.baseURL = url
	}
}

func WithLogger(log *zap.Logger) ConfigOption {
	return func(c *Config) {
		c.log = log.With(zap.String("component", "distr-client"))
	}
}

func WithContextAuth(useContextAuth bool) ConfigOption {
	return func(c *Config) {
		c.useContextAuth = useContextAuth
	}
}

func NewConfig(opts ...ConfigOption) *Config {
	config := Config{
		baseURL:    defaultBaseUrl,
		httpClient: &http.Client{},
		log:        zap.L(),
	}
	for _, opt := range opts {
		opt(&config)
	}
	config.httpClient.Transport = config.roundTripper()
	return &config
}

type Config struct {
	log            *zap.Logger
	baseURL        *url.URL
	token          authkey.Key
	httpClient     *http.Client
	useContextAuth bool
}

func (c *Config) String() string {
	return fmt.Sprintf("client.Config{baseURL: %v, token: %v}", c.baseURL, c.token)
}

func (c *Config) apiUrl(elem ...string) *url.URL {
	return c.baseURL.JoinPath(elem...)
}

func (c *Config) roundTripper() http.RoundTripper {
	rt := tokenRoundTripper{c, c.httpClient.Transport}
	if rt.delegate == nil {
		rt.delegate = http.DefaultTransport
	}
	return &rt
}

type tokenRoundTripper struct {
	*Config
	delegate http.RoundTripper
}

// RoundTrip implements http.RoundTripper.
func (t *tokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Check if we should use context-based auth
	if t.useContextAuth {
		if token, ok := AuthTokenFromContext(req.Context()); ok && token != "" {
			req.Header.Set("Authorization", token)
		}
	} else {
		// Use token from config
		req.Header.Set("Authorization", "AccessToken "+t.token.Serialize())
	}

	if req.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return t.delegate.RoundTrip(req)
}
