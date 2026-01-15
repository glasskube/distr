package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/distr-sh/distr/internal/httpstatus"
	"github.com/distr-sh/distr/internal/types"
	"github.com/google/uuid"
)

func (c *Client) ApplicationVersions(applicationID uuid.UUID) *ApplicationVersions {
	return &ApplicationVersions{config: c.config}
}

type ApplicationVersions struct {
	config        *Config
	applicationID uuid.UUID
}

func (c *ApplicationVersions) url(elem ...string) *url.URL {
	return c.config.apiUrl(append([]string{"api", "v1", "applications", c.applicationID.String(), "versions"}, elem...)...)
}

func (c *ApplicationVersions) Get(ctx context.Context, id uuid.UUID) (*types.ApplicationVersion, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(id.String()).String(), nil)
	if err != nil {
		return nil, err
	}
	return JsonResponse[*types.ApplicationVersion](c.config.httpClient.Do(req))
}

func (c *ApplicationVersions) Create(
	ctx context.Context,
	body types.ApplicationVersion,
) (*types.ApplicationVersion, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url().String(), &buf)
	if err != nil {
		return nil, err
	}
	return JsonResponse[*types.ApplicationVersion](c.config.httpClient.Do(req))
}

func (c *ApplicationVersions) Update(
	ctx context.Context,
	body types.ApplicationVersion,
) (*types.ApplicationVersion, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.url().String(), &buf)
	if err != nil {
		return nil, err
	}
	return JsonResponse[*types.ApplicationVersion](c.config.httpClient.Do(req))
}

func (c *ApplicationVersions) ComposeFile(ctx context.Context, id uuid.UUID) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(id.String(), "compose-file").String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpstatus.CheckStatus(c.config.httpClient.Do(req))
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (c *ApplicationVersions) ValuesFile(ctx context.Context, id uuid.UUID) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(id.String(), "values-file").String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpstatus.CheckStatus(c.config.httpClient.Do(req))
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (c *ApplicationVersions) TemplateFile(ctx context.Context, id uuid.UUID) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(id.String(), "template-file").String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpstatus.CheckStatus(c.config.httpClient.Do(req))
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
