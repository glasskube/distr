package client

import (
	"context"

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
