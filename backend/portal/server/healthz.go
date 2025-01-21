package server

import (
	"context"

	dgrpcserver "github.com/streamingfast/dgrpc/server"
)

func (s *Server) healthzHandler() dgrpcserver.HealthCheck {
	return func(ctx context.Context) (bool, interface{}, error) {
		if !s.isAppReady() {
			return false, nil, nil
		}

		if s.IsTerminating() {
			return false, nil, nil
		}
		return true, nil, nil
	}
}
