package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/distr-sh/distr/api"
	"github.com/distr-sh/distr/internal/httpstatus"
	"github.com/distr-sh/distr/internal/types"
	"github.com/google/uuid"
)

func (c *Client) Applications() *Applications {
	return &Applications{config: c.config}
}

type Applications struct {
	config *Config
}

func (c *Applications) url(elem ...string) *url.URL {
	return c.config.apiUrl(append([]string{"api", "v1", "applications"}, elem...)...)
}

func (c *Applications) List(ctx context.Context) ([]api.ApplicationResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url().String(), nil)
	if err != nil {
		return nil, err
	}
	return JsonResponse[[]api.ApplicationResponse](c.config.httpClient.Do(req))
}

func (c *Applications) Get(ctx context.Context, id uuid.UUID) (*api.ApplicationResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(id.String()).String(), nil)
	if err != nil {
		return nil, err
	}
	return JsonResponse[*api.ApplicationResponse](c.config.httpClient.Do(req))
}

func (c *Applications) Create(ctx context.Context, body types.Application) (*types.Application, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url().String(), &buf)
	if err != nil {
		return nil, err
	}
	return JsonResponse[*types.Application](c.config.httpClient.Do(req))
}

func (c *Applications) Update(ctx context.Context, body types.Application) (*types.Application, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.url(body.ID.String()).String(), &buf)
	if err != nil {
		return nil, err
	}
	return JsonResponse[*types.Application](c.config.httpClient.Do(req))
}

func (c *Applications) Patch(
	ctx context.Context,
	id uuid.UUID,
	body api.PatchApplicationRequest,
) (*types.Application, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.url(id.String()).String(), &buf)
	if err != nil {
		return nil, err
	}
	return JsonResponse[*types.Application](c.config.httpClient.Do(req))
}

func (c *Applications) Delete(ctx context.Context, id uuid.UUID) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url(id.String()).String(), nil)
	if err != nil {
		return err
	}
	_, err = httpstatus.CheckStatus(c.config.httpClient.Do(req))
	return err
}
