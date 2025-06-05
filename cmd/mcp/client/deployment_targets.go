package client

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
)

func (c *Client) DeploymentTargets() *DeploymentTargetsClient {
	return &DeploymentTargetsClient{config: c.config}
}

type DeploymentTargetsClient struct {
	config *Config
}

func (c *DeploymentTargetsClient) List(ctx context.Context) ([]types.DeploymentTargetWithCreatedBy, error) {
	return JsonResponse[[]types.DeploymentTargetWithCreatedBy](
		c.config.httpClient.Get(c.config.apiUrl("api", "v1", "deployment-targets")),
	)
}

func (c *DeploymentTargetsClient) Get(ctx context.Context, id uuid.UUID) (*types.DeploymentTargetWithCreatedBy, error) {
	return JsonResponse[*types.DeploymentTargetWithCreatedBy](
		c.config.httpClient.Get(c.config.apiUrl("api", "v1", "deployment-targets", id.String())),
	)
}

func (c *DeploymentTargetsClient) Create(ctx context.Context, req types.DeploymentTargetWithCreatedBy) (*types.DeploymentTargetWithCreatedBy, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return nil, err
	} else {
		return JsonResponse[*types.DeploymentTargetWithCreatedBy](
			c.config.httpClient.Post(c.config.apiUrl("api", "v1", "deployment-targets"), "application/json", &buf),
		)
	}
}
