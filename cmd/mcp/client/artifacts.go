package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/distr-sh/distr/api"
	"github.com/google/uuid"
)

func (c *Client) Artifacts() *Artifacts {
	return &Artifacts{config: c.config}
}

type Artifacts struct {
	config *Config
}

func (c *Artifacts) url(elem ...string) string {
	return c.config.apiUrl(append([]string{"api", "v1", "artifacts"}, elem...)...).String()
}

// List retrieves all artifacts
func (c *Artifacts) List(ctx context.Context) ([]api.ArtifactsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(), nil)
	if err != nil {
		return nil, err
	}
	return JsonResponse[[]api.ArtifactsResponse](c.config.httpClient.Do(req))
}

// Get retrieves a specific artifact by ID
func (c *Artifacts) Get(ctx context.Context, id uuid.UUID) (*api.ArtifactResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(id.String()), nil)
	if err != nil {
		return nil, err
	}
	return JsonResponse[*api.ArtifactResponse](c.config.httpClient.Do(req))
}

// UpdateImage updates the image for an artifact
func (c *Artifacts) UpdateImage(
	ctx context.Context,
	artifactID uuid.UUID,
	imageID uuid.UUID,
) (*api.ArtifactResponse, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(api.PatchImageRequest{ImageID: imageID}); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.url(artifactID.String(), "image"), &buf)
	if err != nil {
		return nil, err
	}
	return JsonResponse[*api.ArtifactResponse](c.config.httpClient.Do(req))
}
