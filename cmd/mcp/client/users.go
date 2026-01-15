package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/distr-sh/distr/api"
	"github.com/distr-sh/distr/internal/httpstatus"
	"github.com/google/uuid"
)

func (c *Client) Users() *Users {
	return &Users{config: c.config}
}

type Users struct {
	config *Config
}

func (c *Users) url(elem ...string) string {
	return c.config.apiUrl(append([]string{"api", "v1", "user-accounts"}, elem...)...).String()
}

// List retrieves all user accounts
func (c *Users) List(ctx context.Context) ([]api.UserAccountResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(), nil)
	if err != nil {
		return nil, err
	}

	return JsonResponse[[]api.UserAccountResponse](c.config.httpClient.Do(req))
}

// Create creates a new user account
func (c *Users) Create(ctx context.Context, body api.CreateUserAccountRequest) (*api.CreateUserAccountResponse, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url(), &buf)
	if err != nil {
		return nil, err
	}

	return JsonResponse[*api.CreateUserAccountResponse](c.config.httpClient.Do(req))
}

// Delete deletes a user account
func (c *Users) Delete(ctx context.Context, userID uuid.UUID) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url(userID.String()), nil)
	if err != nil {
		return err
	}
	_, err = httpstatus.CheckStatus(c.config.httpClient.Do(req))
	return err
}

// UpdateImage updates the image for a user
func (c *Users) UpdateImage(
	ctx context.Context,
	userID uuid.UUID,
	imageID uuid.UUID,
) (*api.UserAccountResponse, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(api.PatchImageRequest{ImageID: imageID}); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.url(userID.String(), "image"), &buf)
	if err != nil {
		return nil, err
	}

	return JsonResponse[*api.UserAccountResponse](c.config.httpClient.Do(req))
}
