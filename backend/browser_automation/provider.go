package browser_automation

import (
	"context"
)

type CDPInput struct {
	StartURL string
	UseProxy bool
	LiveURL  bool
}

type CDPInfo struct {
	SessionID      string
	WSEndpoint     string
	LiveURL        string
	ReleaseSession ReleaseSession
}

type ReleaseSession func() error

const loginURL = "https://www.reddit.com/login"
const chatURL = "https://chat.reddit.com"
const redditHomePage = "https://www.reddit.com"

type BrowserAutomationProvider interface {
	GetCDPInfo(ctx context.Context, input CDPInput) (*CDPInfo, error)
}
