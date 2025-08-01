package interactions

import (
	"context"
	"github.com/shank318/doota/models"
)

type CDPInfo struct {
	SessionID      string
	WSEndpoint     string
	LiveURL        string
	ReleaseSession ReleaseSession
}

type ReleaseSession func() error

const loginURL = "https://www.reddit.com/login"
const chatURL = "https://chat.reddit.com"

type BrowserAutomation interface {
	StartLogin(ctx context.Context) (*CDPInfo, error)
	WaitAndGetCookies(ctx context.Context, cdp *CDPInfo) (*models.RedditDMLoginConfig, error)
	ValidateCookies(ctx context.Context, cookiesJSON string) (config *models.RedditDMLoginConfig, err error)
	SendDM(ctx context.Context, params DMParams) (cookies []byte, err error)
}
