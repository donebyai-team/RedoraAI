package portal

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/shank318/doota/auth"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
)

func (p *Portal) gethAuthContext(ctx context.Context) (*auth.AuthContext, error) {
	cred, ok := auth.FromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("unauthenticated user access"))

	}
	p.setupLogger(ctx, cred)
	return cred, nil
}

func (p *Portal) setupLogger(ctx context.Context, user *auth.AuthContext) {
	logger := logging.Logger(ctx, p.logger)
	logger = logger.With(zap.String("actor", user.ID), zap.String("actor_org_id", user.OrganizationID))
	logging.WithLogger(ctx, logger)
}
