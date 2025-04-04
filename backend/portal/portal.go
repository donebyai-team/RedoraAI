package portal

import (
	"context"
	"github.com/shank318/doota/agents"
	"github.com/shank318/doota/auth"
	"github.com/shank318/doota/datastore"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"github.com/shank318/doota/portal/server"
	"github.com/shank318/doota/services"
	"github.com/streamingfast/logging"
	"github.com/streamingfast/shutter"
	"go.uber.org/zap"
	"regexp"
)

type Portal struct {
	*shutter.Shutter
	authUsecase         *services.AuthUsecase
	isAppReady          func() bool
	httpListenAddr      string
	corsURLRegexAllow   *regexp.Regexp
	domainWhitelist     []*regexp.Regexp
	db                  datastore.Repository
	config              *pbportal.Config
	logger              *zap.Logger
	tracer              logging.Tracer
	authenticator       *auth.Authenticator
	vanaWebhookHandler  agents.WebhookHandler
	customerCaseService services.CustomerCaseService
}

func New(
	authenticator *auth.Authenticator,
	customerCaseService services.CustomerCaseService,
	vanaWebhookHandler agents.WebhookHandler,
	db datastore.Repository,
	httpListenAddr string,
	corsURLRegexAllow *regexp.Regexp,
	config *pbportal.Config,
	domainWhitelist []*regexp.Regexp,
	isAppReady func() bool,
	logger *zap.Logger,
	tracer logging.Tracer,
) *Portal {
	return &Portal{
		Shutter:             shutter.New(),
		config:              config,
		authenticator:       authenticator,
		vanaWebhookHandler:  vanaWebhookHandler,
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

	s.Run(p, p.UpdateCallStatusHandler, p.EndConversationHandler, p.HandleBatchAdmin)
	return nil
}
