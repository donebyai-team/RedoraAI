package portal

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"connectrpc.com/connect"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"github.com/shank318/doota/portal/state"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
)

func randomString(n int) (string, error) {
	data := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// OAuthAuthorize implements pbportalconnect.PortalServiceHandler.
func (p *Portal) OauthAuthorize(ctx context.Context, req *connect.Request[pbportal.OauthAuthorizeRequest]) (*connect.Response[pbportal.OauthAuthorizeResponse], error) {
	// 1. Validate the auth credentials and authorization type
	//actor, err := p.gethAuthContext(ctx)
	//if err != nil {
	//	return nil, err
	//}

	logger := logging.Logger(ctx, p.logger)
	logger.Info("fetching OAuth url with state assignment", zap.String("redirect_uri", req.Msg.RedirectUrl))

	// 2. Create a state object, store in redis as SetState
	stateHash, err := randomString(15)
	if err != nil {
		return nil, fmt.Errorf("unable to generate state: %w", err)
	}

	nonce, err := randomString(15)
	if err != nil {
		return nil, fmt.Errorf("unable to generate nonce: %w", err)
	}

	state := &state.State{
		Hash:            stateHash,
		Nonce:           nonce,
		RedirectUri:     req.Msg.RedirectUrl,
		IntegrationType: req.Msg.IntegrationType,
		ExpiresAt:       time.Now().Add(5 * time.Minute),
	}

	logger.Debug("storing state", zap.Reflect("state", state))
	p.authStateStore.SetState(state)

	authorizeURL := ""
	switch req.Msg.IntegrationType {
	case pbportal.IntegrationType_INTEGRATION_TYPE_REDDIT:
		authorizeURL = p.redditOauthClient.GetAuthURL(state.Hash)
	case pbportal.IntegrationType_INTEGRATION_TYPE_GOOGLE:
		authorizeURL = p.googleOauthClient.AuthorizeURL(state.Hash)
	default:
		return nil, fmt.Errorf("unknown integration: %s", req.Msg.IntegrationType)
	}

	return connect.NewResponse(&pbportal.OauthAuthorizeResponse{
		AuthorizeUrl: authorizeURL,
	}), nil
}
