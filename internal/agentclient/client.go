package agentclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/glasskube/distr/internal/agentclient/useragent"
	"github.com/glasskube/distr/internal/buildconfig"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"

	"github.com/glasskube/distr/api"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type clientData struct {
	authTarget       string
	authSecret       string
	loginEndpoint    string
	manifestEndpoint string
	resourceEndpoint string
	statusEndpoint   string
}

type Client struct {
	clientData
	httpClient *http.Client
	logger     *zap.Logger
	token      jwt.Token
	rawToken   string
}

func (c *Client) Resource(ctx context.Context) (*api.AgentResource, error) {
	var result api.AgentResource
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

func (c *Client) StatusWithError(ctx context.Context, revisionID uuid.UUID, message string, err error) error {
	statusType := types.DeploymentStatusTypeOK
	if err != nil {
		statusType = types.DeploymentStatusTypeError
		message = err.Error()
	}
	return c.Status(ctx, revisionID, statusType, message)
}

func (c *Client) Status(
	ctx context.Context,
	revisionID uuid.UUID,
	statusType types.DeploymentStatusType,
	message string,
) error {
	deploymentStatus := api.AgentDeploymentStatus{
		RevisionID: revisionID,
		Message:    message,
		Type:       statusType,
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

func (c *Client) Logs(ctx context.Context, logs []api.LogRecord) error {
	log.Printf("pushing logs: %v", logs)
	return nil
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

func (c *Client) ClearToken() {
	c.token = nil
	c.rawToken = ""
}

func (c *Client) RawToken() string {
	return c.rawToken
}

func (c *Client) doAuthenticated(ctx context.Context, r *http.Request) (*http.Response, error) {
	if resp, err := c.doAuthenticatedNoRetry(ctx, r); resp == nil || resp.StatusCode != 401 {
		return resp, err
	} else {
		c.logger.Warn("got 401 response, try to regenerate token")
		c.ClearToken()
		resp, err1 := c.doAuthenticatedNoRetry(ctx, r)
		if err1 != nil {
			return resp, multierr.Append(err, err1)
		} else {
			return resp, nil
		}
	}
}

func (c *Client) doAuthenticatedNoRetry(ctx context.Context, r *http.Request) (*http.Response, error) {
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

func (c *Client) ReloadFromEnv() (changed bool, err error) {
	var d clientData
	if d.authTarget, err = readEnvVar("DISTR_TARGET_ID"); err != nil {
		return
	} else if d.authSecret, err = readEnvVar("DISTR_TARGET_SECRET"); err != nil {
		return
	} else if d.loginEndpoint, err = readEnvVar("DISTR_LOGIN_ENDPOINT"); err != nil {
		return
	} else if d.manifestEndpoint, err = readEnvVar("DISTR_MANIFEST_ENDPOINT"); err != nil {
		return
	} else if d.resourceEndpoint, err = readEnvVar("DISTR_RESOURCE_ENDPOINT"); err != nil {
		return
	} else if d.statusEndpoint, err = readEnvVar("DISTR_STATUS_ENDPOINT"); err != nil {
		return
	} else {
		changed = c.clientData != d
		if changed {
			c.clientData = d
			c.ClearToken()
		}
		return
	}
}

func NewFromEnv(logger *zap.Logger) (*Client, error) {
	client := Client{
		httpClient: &http.Client{},
		logger:     logger,
	}
	if _, err := client.ReloadFromEnv(); err != nil {
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
