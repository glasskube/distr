package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/distr-sh/distr/api"
	"github.com/distr-sh/distr/internal/httpstatus"
	"github.com/distr-sh/distr/internal/types"
	"github.com/google/uuid"
)

func (c *Client) Deployments() *Deployments {
	return &Deployments{config: c.config}
}

type Deployments struct {
	config *Config
}

func (c *Deployments) url(elem ...string) *url.URL {
	return c.config.apiUrl(append([]string{"api", "v1", "deployments"}, elem...)...)
}

func (c *Deployments) Put(ctx context.Context, req api.DeploymentRequest) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return err
	} else if req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.url().String(), &buf); err != nil {
		return err
	} else {
		_, err := httpstatus.CheckStatus(c.config.httpClient.Do(req))
		return err
	}
}

func (c *Deployments) Patch(
	ctx context.Context,
	id uuid.UUID,
	req api.PatchDeploymentRequest,
) (*types.Deployment, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return nil, err
	} else if req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.url(id.String()).String(), &buf); err != nil {
		return nil, err
	} else {
		return JsonResponse[*types.Deployment](c.config.httpClient.Do(req))
	}
}

func (c *Deployments) Delete(ctx context.Context, id uuid.UUID) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url(id.String()).String(), nil)
	if err != nil {
		return err
	}
	_, err = httpstatus.CheckStatus(c.config.httpClient.Do(req))
	return err
}

type TimeseriesResourceOptions struct {
	Limit  *int64
	Before *time.Time
	After  *time.Time
}

func (o *TimeseriesResourceOptions) AsURLValues() url.Values {
	if o == nil {
		return nil
	}
	v := url.Values{}
	if o.Limit != nil {
		v.Add("limit", strconv.FormatInt(*o.Limit, 10))
	}
	if o.Before != nil {
		v.Add("before", o.Before.Format(time.RFC3339Nano))
	}
	if o.After != nil {
		v.Add("before", o.After.Format(time.RFC3339Nano))
	}
	return v
}

func (c *Deployments) Status(
	ctx context.Context,
	id uuid.UUID,
	options *TimeseriesResourceOptions,
) ([]types.DeploymentRevisionStatus, error) {
	url := c.url(id.String(), "status")
	url.RawQuery = options.AsURLValues().Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}
	return JsonResponse[[]types.DeploymentRevisionStatus](c.config.httpClient.Do(req))
}

func (c *Deployments) Logs(
	ctx context.Context,
	id uuid.UUID,
	resource string,
	options *TimeseriesResourceOptions,
) ([]api.DeploymentLogRecord, error) {
	url := c.url(id.String(), "logs")
	url.RawQuery = options.AsURLValues().Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}
	return JsonResponse[[]api.DeploymentLogRecord](c.config.httpClient.Do(req))
}

func (c *Deployments) LogResources(ctx context.Context, id uuid.UUID) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(id.String(), "logs", "resources").String(), nil)
	if err != nil {
		return nil, err
	}
	return JsonResponse[[]string](c.config.httpClient.Do(req))
}
