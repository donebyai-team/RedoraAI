package browser_automation

import (
	"context"
	"strings"
)

type CDPInput struct {
	StartURL          string
	UseProxy          bool
	LiveURL           bool
	IsWarmUp          bool
	Alpha2CountryCode string
}

const defaultAlpha2CountryCode = "US"

func (c CDPInput) GetCountryCode() string {
	if code := strings.TrimSpace(c.Alpha2CountryCode); code != "" {
		return code
	}
	return defaultAlpha2CountryCode
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
