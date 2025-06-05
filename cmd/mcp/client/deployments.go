package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/httpstatus"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
)

func (c *Client) Deployments() *Deployments {
	return &Deployments{config: c.config}
}

type Deployments struct {
	config *Config
}

func (c *Deployments) url(elem ...string) string {
	return c.config.apiUrl(append([]string{"api", "v1", "deployments"}, elem...)...)
}

func (c *Deployments) Put(ctx context.Context, req api.DeploymentRequest) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return err
	} else if req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.url(), &buf); err != nil {
		return err
	} else {
		_, err := httpstatus.CheckStatus(c.config.httpClient.Do(req))
		return err
	}
}

func (c *Deployments) Patch(ctx context.Context, id uuid.UUID, req api.PatchDeploymentRequest) (*types.Deployment, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return nil, err
	} else if req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.url(id.String()), &buf); err != nil {
		return nil, err
	} else {
		return JsonResponse[*types.Deployment](c.config.httpClient.Do(req))
	}
}

func (c *Deployments) Delete(ctx context.Context, id uuid.UUID) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url(id.String()), nil)
	if err != nil {
		return err
	}
	_, err = httpstatus.CheckStatus(c.config.httpClient.Do(req))
	return err
}

func (c *Deployments) Status(ctx context.Context, id uuid.UUID) ([]types.DeploymentRevisionStatus, error) {
	return nil, nil
}

func (c *Deployments) LogResources(ctx context.Context, id uuid.UUID) ([]string, error) {
	return nil, nil
}

func (c *Deployments) Logs(ctx context.Context, id uuid.UUID, resource string) ([]api.DeploymentLogRecord, error) {
	return nil, nil
}
