package portal

import (
	"context"
	state2 "github.com/shank318/doota/agents/state"
	"github.com/shank318/doota/ai"
	"github.com/shank318/doota/integrations/reddit"
	"github.com/shank318/doota/portal/state"
	"regexp"

	"github.com/shank318/doota/agents"
	"github.com/shank318/doota/auth"
	"github.com/shank318/doota/datastore"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"github.com/shank318/doota/portal/server"
	"github.com/shank318/doota/services"
	"github.com/streamingfast/logging"
	"github.com/streamingfast/shutter"
	"go.uber.org/zap"
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
	keywordService      services.KeywordService
	authStateStore      state.AuthStateStore
	redditOauthClient   *reddit.OauthClient
	aiClient            *ai.Client
	cache               state2.ConversationState
}

func New(
	aiClient *ai.Client,
	redditOauthClient *reddit.OauthClient,
	authenticator *auth.Authenticator,
	authStateStore state.AuthStateStore,
	customerCaseService services.CustomerCaseService,
	authUsecase *services.AuthUsecase,
	keywordService services.KeywordService,
	vanaWebhookHandler agents.WebhookHandler,
	db datastore.Repository,
	cache state2.ConversationState,
	httpListenAddr string,
	corsURLRegexAllow *regexp.Regexp,
	config *pbportal.Config,
	domainWhitelist []*regexp.Regexp,
	isAppReady func() bool,
	logger *zap.Logger,
	tracer logging.Tracer,
) *Portal {
	return &Portal{
		aiClient:            aiClient,
		redditOauthClient:   redditOauthClient,
		authStateStore:      authStateStore,
		authUsecase:         authUsecase,
		Shutter:             shutter.New(),
		config:              config,
		authenticator:       authenticator,
		vanaWebhookHandler:  vanaWebhookHandler,
		customerCaseService: customerCaseService,
		keywordService:      keywordService,
		db:                  db,
		cache:               cache,
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
