package portal

import (
	"context"
	"github.com/shank318/doota/auth"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/portal/server"
	"github.com/shank318/doota/services"
	"github.com/streamingfast/logging"
	"github.com/streamingfast/shutter"
	"go.uber.org/zap"
	"regexp"
)

type Portal struct {
	*shutter.Shutter
	isAppReady          func() bool
	httpListenAddr      string
	corsURLRegexAllow   *regexp.Regexp
	domainWhitelist     []*regexp.Regexp
	db                  datastore.Repository
	logger              *zap.Logger
	tracer              logging.Tracer
	authenticator       *auth.Authenticator
	customerCaseService services.CustomerCaseService
}

func New(
	authenticator *auth.Authenticator,
	customerCaseService services.CustomerCaseService,
	db datastore.Repository,
	httpListenAddr string,
	corsURLRegexAllow *regexp.Regexp,
	domainWhitelist []*regexp.Regexp,
	isAppReady func() bool,
	logger *zap.Logger,
	tracer logging.Tracer,
) *Portal {
	return &Portal{
		Shutter:             shutter.New(),
		authenticator:       authenticator,
		customerCaseService: customerCaseService,
		db:                  db,
		httpListenAddr:      httpListenAddr,
		corsURLRegexAllow:   corsURLRegexAllow,
		domainWhitelist:     domainWhitelist,
		isAppReady:          isAppReady,
		logger:              logger.Named("portal"),
		tracer:              tracer,
	}
}

func (p *Portal) Run(ctx context.Context) error {
	p.logger.Info("starting portal server", zap.String("http_listen_addr", p.httpListenAddr))
	s := server.New(p.httpListenAddr, p.authenticator, p.corsURLRegexAllow, p.isAppReady, p.logger)
	p.OnTerminating(func(_ error) {
		s.Shutdown(nil)
	})

	s.Run(p)
	return nil
}
