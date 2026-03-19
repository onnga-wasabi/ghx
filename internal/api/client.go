package api

import (
	"context"
	"net/http"

	"github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"
)

type Client struct {
	GH    *github.Client
	Token string
}

func NewClient(token string) *Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	return &Client{
		GH:    github.NewClient(tc),
		Token: token,
	}
}

func (c *Client) HTTPClient() *http.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: c.Token})
	return oauth2.NewClient(context.Background(), ts)
}
