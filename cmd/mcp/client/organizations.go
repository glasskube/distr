package client

import (
	"context"
	"net/http"

	"github.com/distr-sh/distr/internal/types"
)

func (c *Client) Organizations() *Organizations {
	return &Organizations{config: c.config}
}

type Organizations struct {
	config *Config
}

func (c *Organizations) url(elem ...string) string {
	return c.config.apiUrl(append([]string{"api", "v1", "organizations"}, elem...)...).String()
}

// List retrieves all organizations for the current user
func (c *Organizations) List(ctx context.Context) ([]types.OrganizationWithUserRole, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(), nil)
	if err != nil {
		return nil, err
	}

	return JsonResponse[[]types.OrganizationWithUserRole](c.config.httpClient.Do(req))
}
