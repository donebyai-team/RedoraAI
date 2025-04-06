package reddit

import (
	"context"
	"go.uber.org/zap"
)

type Client struct {
	// add a reddit client (go package)
	logger *zap.Logger
}

func New(logger *zap.Logger) *Client {
	return &Client{logger: logger}
}

func (c *Client) GetPosts(ctx context.Context, subRedditID string) {

}

func (c *Client) GetUser(ctx context.Context) {

}
