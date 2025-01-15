package agentclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/glasskube/cloud/internal/types"

	"github.com/glasskube/cloud/api"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"go.uber.org/zap"
)

type Client struct {
	AuthTarget string
	AuthSecret string

	LoginEndpoint    string
	ResourceEndpoint string
	StatusEndpoint   string

	httpClient *http.Client
	logger     *zap.Logger
	token      jwt.Token
	rawToken   string
}

func (c *Client) Resource(ctx context.Context) (*api.DockerAgentResource, error) {
	var result api.DockerAgentResource
	if req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.ResourceEndpoint, nil); err != nil {
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

func (c *Client) KubernetesResource(ctx context.Context) (*api.KubernetesAgentResource, error) {
	var result api.KubernetesAgentResource
	if req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.ResourceEndpoint, nil); err != nil {
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

func (c *Client) Status(ctx context.Context, revisionID string, status string, error error) error {
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
	} else if req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.StatusEndpoint, &buf); err != nil {
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.LoginEndpoint, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.AuthTarget, c.AuthSecret)
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
	return checkStatus(c.httpClient.Do(r))
}

func NewFromEnv(logger *zap.Logger) (*Client, error) {
	client := Client{
		httpClient: &http.Client{},
		logger:     logger,
	}
	var err error
	if client.AuthTarget, err = readEnvVar("GK_TARGET_ID"); err != nil {
		return nil, err
	}
	if client.AuthSecret, err = readEnvVar("GK_TARGET_SECRET"); err != nil {
		return nil, err
	}
	if client.LoginEndpoint, err = readEnvVar("GK_LOGIN_ENDPOINT"); err != nil {
		return nil, err
	}
	if client.ResourceEndpoint, err = readEnvVar("GK_RESOURCE_ENDPOINT"); err != nil {
		return nil, err
	}
	if client.StatusEndpoint, err = readEnvVar("GK_STATUS_ENDPOINT"); err != nil {
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
