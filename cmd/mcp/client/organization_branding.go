package client

import (
	"context"
	"net/http"

	"github.com/distr-sh/distr/internal/types"
)

func (c *Organization) Branding() *OrganizationBranding {
	return &OrganizationBranding{config: c.config}
}

type OrganizationBranding struct {
	config *Config
}

func (c *OrganizationBranding) url(elem ...string) string {
	return c.config.apiUrl(append([]string{"api", "v1", "organization", "branding"}, elem...)...).String()
}

func (c *OrganizationBranding) Get(ctx context.Context) (*types.OrganizationBranding, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(), nil)
	if err != nil {
		return nil, err
	}

	return JsonResponse[*types.OrganizationBranding](c.config.httpClient.Do(req))
}
