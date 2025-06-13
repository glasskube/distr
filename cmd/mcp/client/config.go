package client

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/glasskube/distr/internal/authkey"
	"github.com/glasskube/distr/internal/util"
	"go.uber.org/zap"
)

var defaultBaseUrl = util.Require(url.Parse("https://app.distr.sh/"))

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
	log        *zap.Logger
	baseURL    *url.URL
	token      authkey.Key
	httpClient *http.Client
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
	req.Header.Set("Authorization", "AccessToken "+t.token.Serialize())
	if req.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return t.delegate.RoundTrip(req)
}
