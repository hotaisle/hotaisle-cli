package api

import (
	"hotaisle-cli/client"
)

const DefaultHost = "admin.hotaisle.app"
const DefaultBaseURL = "https://" + DefaultHost + "/api"

type Client struct {
	Api *client.Client
}

func NewClient(token string, version string) *Client {
	return &Client{
		Api: client.NewClient(
			client.WithBaseURL(DefaultBaseURL),
			client.WithToken(token),
			client.WithUserAgent("hotaisle/"+version),
		),
	}
}
