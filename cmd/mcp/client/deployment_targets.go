package client

import (
	"context"
	"fmt"
	"net/http"

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
	if req, err := c.config.NewAuthenticatedRequest(ctx, http.MethodGet, "api/v1/deployment-targets", nil); err != nil {
		return nil, err
	} else {
		return Do[[]types.DeploymentTargetWithCreatedBy](c.config, req)
	}
}

func (c *DeploymentTargetsClient) Get(ctx context.Context, id uuid.UUID) (*types.DeploymentTargetWithCreatedBy, error) {
	if req, err := c.config.NewAuthenticatedRequest(
		ctx,
		http.MethodGet,
		fmt.Sprintf("api/v1/deployment-targets/%v", id),
		nil,
	); err != nil {
		return nil, err
	} else {
		return Do[*types.DeploymentTargetWithCreatedBy](c.config, req)
	}
}
