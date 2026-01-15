package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/distr-sh/distr/internal/types"
)

func (c *Client) Organization() *Organization {
	return &Organization{config: c.config}
}

type Organization struct {
	config *Config
}

func (c *Organization) url(elem ...string) string {
	return c.config.apiUrl(append([]string{"api", "v1", "organization"}, elem...)...).String()
}

// Current retrieves the current organization
func (c *Organization) Current(ctx context.Context) (*types.Organization, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("current"), nil)
	if err != nil {
		return nil, err
	}

	return JsonResponse[*types.Organization](c.config.httpClient.Do(req))
}

// Create creates a new organization
func (c *Organization) Create(ctx context.Context, org *types.Organization) (*types.OrganizationWithUserRole, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(org); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url(), &buf)
	if err != nil {
		return nil, err
	}

	return JsonResponse[*types.OrganizationWithUserRole](c.config.httpClient.Do(req))
}

// Update updates an existing organization
func (c *Organization) Update(ctx context.Context, org *types.Organization) (*types.Organization, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(org); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.url(), &buf)
	if err != nil {
		return nil, err
	}

	return JsonResponse[*types.Organization](c.config.httpClient.Do(req))
}
