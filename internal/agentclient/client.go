package agentclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/glasskube/distr/internal/agentclient/useragent"
	"github.com/glasskube/distr/internal/buildconfig"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"

	"github.com/glasskube/distr/api"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"go.uber.org/zap"
)

type Client struct {
	authTarget string
	authSecret string

	loginEndpoint    string
	manifestEndpoint string
	resourceEndpoint string
	statusEndpoint   string

	httpClient *http.Client
	logger     *zap.Logger
	token      jwt.Token
	rawToken   string
}

func (c *Client) DockerResource(ctx context.Context) (*api.DockerAgentResource, error) {
	return resource[api.DockerAgentResource](ctx, c)
}

func (c *Client) KubernetesResource(ctx context.Context) (*api.KubernetesAgentResource, error) {
	return resource[api.KubernetesAgentResource](ctx, c)
}

func resource[T any](ctx context.Context, c *Client) (*T, error) {
	var result T
	if req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.resourceEndpoint, nil); err != nil {
		return nil, err
	} else {
		req.Header.Set("Content-Type", "application/json")
		if resp, err := c.doAuthenticated(ctx, req); err != nil {
			return nil, err
		} else if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		} else {
			return &result, nil
		}
	}
}

func (c *Client) Manifest(ctx context.Context) ([]byte, error) {
	if req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.manifestEndpoint, nil); err != nil {
		return nil, err
	} else if resp, err := c.doAuthenticated(ctx, req); err != nil {
		return nil, err
	} else if data, err := io.ReadAll(resp.Body); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

func (c *Client) Status(ctx context.Context, revisionID uuid.UUID, status string, error error) error {
	deploymentStatus := api.AgentDeploymentStatus{
		RevisionID: revisionID,
	}
	if error != nil {
		deploymentStatus.Type = types.DeploymentStatusTypeError
		deploymentStatus.Message = error.Error()
	} else {
		deploymentStatus.Type = types.DeploymentStatusTypeOK
		deploymentStatus.Message = status
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(deploymentStatus); err != nil {
		return err
	} else if req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.statusEndpoint, &buf); err != nil {
		return err
	} else {
		req.Header.Set("Content-Type", "application/json")
		if _, err := c.doAuthenticated(ctx, req); err != nil {
			return err
		} else {
			return nil
		}
	}
}

func (c *Client) Login(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.loginEndpoint, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.authTarget, c.authSecret)
	if resp, err := c.do(req); err != nil {
		return err
	} else {
		var loginResponse api.AuthLoginResponse
		if err := json.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
			return err
		}
		if parsedToken, err := jwt.Parse([]byte(loginResponse.Token), jwt.WithVerify(false)); err != nil {
			return err
		} else {
			c.rawToken = loginResponse.Token
			c.token = parsedToken
			return nil
		}
	}
}

func (c *Client) EnsureToken(ctx context.Context) error {
	if c.HasTokenExpiredAfter(time.Now().Add(30 * time.Second)) {
		c.logger.Info("token has expired or is about to expire")
		if err := c.Login(ctx); err != nil {
			if c.HasTokenExpired() {
				return err
			} else {
				c.logger.Warn("token refresh failed but previous token is still valid", zap.Error(err))
			}
		} else {
			c.logger.Info("token refreshed")
		}
	}
	return nil
}

func (c *Client) HasTokenExpired() bool {
	return c.HasTokenExpiredAfter(time.Now())
}

func (c *Client) HasTokenExpiredAfter(t time.Time) bool {
	return c.token == nil || c.token.Expiration().Before(t)
}

func (c *Client) doAuthenticated(ctx context.Context, r *http.Request) (*http.Response, error) {
	if err := c.EnsureToken(ctx); err != nil {
		return nil, err
	} else {
		r.Header.Set("Authorization", "Bearer "+c.rawToken)
		return c.do(r)
	}
}

func (c *Client) do(r *http.Request) (*http.Response, error) {
	r.Header.Set("User-Agent", fmt.Sprintf("%v/%v", useragent.DistrAgentUserAgent, buildconfig.Version()))
	return checkStatus(c.httpClient.Do(r))
}

func NewFromEnv(logger *zap.Logger) (*Client, error) {
	client := Client{
		httpClient: &http.Client{},
		logger:     logger,
	}
	var err error
	if client.authTarget, err = readEnvVar("DISTR_TARGET_ID"); err != nil {
		return nil, err
	}
	if client.authSecret, err = readEnvVar("DISTR_TARGET_SECRET"); err != nil {
		return nil, err
	}
	if client.loginEndpoint, err = readEnvVar("DISTR_LOGIN_ENDPOINT"); err != nil {
		return nil, err
	}
	if client.manifestEndpoint, err = readEnvVar("DISTR_MANIFEST_ENDPOINT"); err != nil {
		return nil, err
	}
	if client.resourceEndpoint, err = readEnvVar("DISTR_RESOURCE_ENDPOINT"); err != nil {
		return nil, err
	}
	if client.statusEndpoint, err = readEnvVar("DISTR_STATUS_ENDPOINT"); err != nil {
		return nil, err
	}
	return &client, nil
}

func readEnvVar(key string) (string, error) {
	if value, ok := os.LookupEnv(key); ok {
		return value, nil
	} else {
		return "", fmt.Errorf("missing environment variable: %v", key)
	}
}
