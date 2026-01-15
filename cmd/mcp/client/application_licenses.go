package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/distr-sh/distr/internal/httpstatus"
	"github.com/distr-sh/distr/internal/types"
	"github.com/google/uuid"
)

func (c *Client) ApplicationLicenses() *ApplicationLicenses {
	return &ApplicationLicenses{config: c.config}
}

type ApplicationLicenses struct {
	config *Config
}

func (c *ApplicationLicenses) url(elem ...string) string {
	return c.config.apiUrl(append([]string{"api", "v1", "application-licenses"}, elem...)...).String()
}

// List retrieves all application licenses
func (c *ApplicationLicenses) List(ctx context.Context) ([]types.ApplicationLicense, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(), nil)
	if err != nil {
		return nil, err
	}
	return JsonResponse[[]types.ApplicationLicense](c.config.httpClient.Do(req))
}

// Get retrieves a specific application license by ID
func (c *ApplicationLicenses) Get(ctx context.Context, id uuid.UUID) (*types.ApplicationLicense, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(id.String()), nil)
	if err != nil {
		return nil, err
	}
	return JsonResponse[*types.ApplicationLicense](c.config.httpClient.Do(req))
}

// Create creates a new application license
func (c *ApplicationLicenses) Create(
	ctx context.Context,
	license *types.ApplicationLicenseWithVersions,
) (*types.ApplicationLicense, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(license); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url(), &buf)
	if err != nil {
		return nil, err
	}
	return JsonResponse[*types.ApplicationLicense](c.config.httpClient.Do(req))
}

// Update updates an existing application license
func (c *ApplicationLicenses) Update(
	ctx context.Context,
	license *types.ApplicationLicenseWithVersions,
) (*types.ApplicationLicense, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(license); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.url(license.ID.String()), &buf)
	if err != nil {
		return nil, err
	}
	return JsonResponse[*types.ApplicationLicense](c.config.httpClient.Do(req))
}

// Delete deletes an application license
func (c *ApplicationLicenses) Delete(ctx context.Context, id uuid.UUID) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url(id.String()), nil)
	if err != nil {
		return err
	}
	_, err = httpstatus.CheckStatus(c.config.httpClient.Do(req))
	return err
}
