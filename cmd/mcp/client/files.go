package client

import (
	"context"
	"io"
	"net/http"

	"github.com/distr-sh/distr/internal/httpstatus"
	"github.com/distr-sh/distr/internal/types"
	"github.com/google/uuid"
)

func (c *Client) Files() *Files {
	return &Files{config: c.config}
}

type Files struct {
	config *Config
}

func (c *Files) url(elem ...string) string {
	return c.config.apiUrl(append([]string{"api", "v1", "files"}, elem...)...).String()
}

// Get retrieves a file by ID
func (c *Files) Get(ctx context.Context, id uuid.UUID) (*types.File, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(id.String()), nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpstatus.CheckStatus(c.config.httpClient.Do(req))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &types.File{
		ID:          id,
		ContentType: resp.Header.Get("Content-Type"),
		FileName:    resp.Header.Get("Content-Disposition"),
		Data:        data,
	}, nil
}

// Delete deletes a file by ID
func (c *Files) Delete(ctx context.Context, id uuid.UUID) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url(id.String()), nil)
	if err != nil {
		return err
	}
	_, err = httpstatus.CheckStatus(c.config.httpClient.Do(req))
	return err
}
