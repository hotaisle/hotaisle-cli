package api

import (
	"hotaisle-cli/client"
)

type Client struct {
	Api *client.Client
}

func NewClient(token string, version string, opts ...client.Option) *Client {
	// Prepare the default options
	defaultOpts := []client.Option{
		client.WithBaseURL(client.DefaultBaseURL),
		client.WithToken(token),
		client.WithUserAgent("hotaisle/" + version),
	}

	// Append any additional options (like WithHTTPClient for testing)
	allOpts := append(defaultOpts, opts...)

	return &Client{
		Api: client.NewClient(allOpts...),
	}
}
