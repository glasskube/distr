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

func (c *Client) DeploymentTargets() *DeploymentTargets {
	return &DeploymentTargets{config: c.config}
}

type DeploymentTargets struct {
	config *Config
}

func (c *DeploymentTargets) url(elem ...string) string {
	return c.config.apiUrl(append([]string{"api", "v1", "deployment-targets"}, elem...)...).String()
}

func (c *DeploymentTargets) List(ctx context.Context) ([]types.DeploymentTargetWithCreatedBy, error) {
	return JsonResponse[[]types.DeploymentTargetWithCreatedBy](
		c.config.httpClient.Get(c.url()),
	)
}

func (c *DeploymentTargets) Get(ctx context.Context, id uuid.UUID) (*types.DeploymentTargetWithCreatedBy, error) {
	return JsonResponse[*types.DeploymentTargetWithCreatedBy](
		c.config.httpClient.Get(c.url(id.String())),
	)
}

func (c *DeploymentTargets) Create(
	ctx context.Context,
	req types.DeploymentTargetWithCreatedBy,
) (*types.DeploymentTargetWithCreatedBy, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return nil, err
	} else {
		return JsonResponse[*types.DeploymentTargetWithCreatedBy](
			c.config.httpClient.Post(c.url(), "application/json", &buf),
		)
	}
}

func (c *DeploymentTargets) Update(
	ctx context.Context,
	req types.DeploymentTargetWithCreatedBy,
) (*types.DeploymentTargetWithCreatedBy, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return nil, err
	} else {
		return JsonResponse[*types.DeploymentTargetWithCreatedBy](
			c.config.httpClient.Post(
				c.url(req.ID.String()),
				"application/json",
				&buf,
			),
		)
	}
}

func (c *DeploymentTargets) Delete(ctx context.Context, id uuid.UUID) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url(id.String()), nil)
	if err != nil {
		return err
	}
	_, err = httpstatus.CheckStatus(c.config.httpClient.Do(req))
	return err
}

func (c *DeploymentTargets) Connect(
	ctx context.Context,
	id uuid.UUID,
) (*api.DeploymentTargetAccessTokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url(id.String(), "access-request"), nil)
	if err != nil {
		return nil, err
	}
	return JsonResponse[*api.DeploymentTargetAccessTokenResponse](c.config.httpClient.Do(req))
}
