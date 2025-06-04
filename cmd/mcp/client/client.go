package client

type Client struct {
	config *Config
}

func NewClient(config *Config) *Client {
	return &Client{config}
}
