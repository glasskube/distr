package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

type Config struct {
	BaseUrl    *url.URL
	Token      string
	HttpClient *http.Client
}

func (c *Config) NewAuthenticatedRequest(
	ctx context.Context,
	method string,
	path string,
	body io.Reader,
) (*http.Request, error) {
	url := c.BaseUrl.JoinPath(path).String()
	if request, err := http.NewRequestWithContext(ctx, method, url, body); err != nil {
		return nil, err
	} else {
		request.Header.Add("AccessToken", c.Token)
		return request, nil
	}
}

func Do[T any](c *Config, req *http.Request) (T, error) {
	var result T
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}
