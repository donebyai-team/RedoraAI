package app

import (
	"context"
	"fmt"
	"github.com/shank318/doota/agents/state"
	"github.com/shank318/doota/ai"
	"github.com/shank318/doota/auth"
	"github.com/shank318/doota/auth/crypto"
	"github.com/shank318/doota/datastore"
	google2 "github.com/shank318/doota/integrations/google"
	"github.com/shank318/doota/models"
	"github.com/streamingfast/dstore"
	"golang.org/x/oauth2"
	"regexp"
	"time"

	"github.com/streamingfast/logging"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	Endpoint     oauth2.Endpoint
}

type DependenciesBuilder struct {
	PGDSN              string
	KMSKeyPath         string
	CorsURLRegexAllow  string
	AttachmentStoreURL string
	PubsubGCPProject   string
	Processor          bool
	AIConfig           *AIConfig
	ConversationState  *conversationState
	GoogleConfig       *GoogleConfig
	dig                *dig.Container
}

func NewDependenciesBuilder() *DependenciesBuilder {
	return &DependenciesBuilder{
		dig: dig.New(),
	}
}

type conversationState struct {
	redisAddr         string
	phoneCallStateTTL time.Duration
	namespace, prefix string
}

type AIConfig struct {
	OpenAIKey            string
	OpenAIOrganization   string
	OpenAIDebugLogsStore string
	LangsmithApiKey      string
	LangsmithProject     string
	DefaultLLMModel      models.LLMModel
	AdvanceLLMModel      models.LLMModel
}

func (b *DependenciesBuilder) mustProvide(constructor interface{}) {
	if err := b.dig.Provide(constructor); err != nil {
		panic(fmt.Errorf("failed to register provider: %w", err))
	}
}

func (b *DependenciesBuilder) WithDataStore(pgDSN string) *DependenciesBuilder {
	b.mustProvide(func() PostgresDSNString { return PostgresDSNString(pgDSN) })
	b.PGDSN = pgDSN
	return b
}

func (b *DependenciesBuilder) WithGoogle(clientId, clientSecret, redirectUrl string) *DependenciesBuilder {
	b.GoogleConfig = &GoogleConfig{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  fmt.Sprintf("%s/login", redirectUrl),
	}
	return b
}

func (b *DependenciesBuilder) WithConversationState(phoneCallStateTTL time.Duration, redisAddr, namespace, prefix string) *DependenciesBuilder {
	b.ConversationState = &conversationState{
		redisAddr,
		phoneCallStateTTL,
		namespace, prefix,
	}
	return b
}

func (b *DependenciesBuilder) WithAI(defaultLLMModel, advanceLLMModel models.LLMModel, openAIKey string, openAIDebugLogsStore string, langsmithApiKey string, langsmithProject string) *DependenciesBuilder {
	b.AIConfig = &AIConfig{
		OpenAIKey:            openAIKey,
		OpenAIDebugLogsStore: openAIDebugLogsStore,
		LangsmithApiKey:      langsmithApiKey,
		LangsmithProject:     langsmithProject,
		DefaultLLMModel:      defaultLLMModel,
		AdvanceLLMModel:      advanceLLMModel,
	}
	return b
}

func (b *DependenciesBuilder) WithKMSKeyPath(kmsKeyPath string) *DependenciesBuilder {
	b.KMSKeyPath = kmsKeyPath
	return b
}

func (b *DependenciesBuilder) WithCORSURLRegexAllow(corsURLRegexAllow string) *DependenciesBuilder {
	b.CorsURLRegexAllow = corsURLRegexAllow
	return b
}

func (b *DependenciesBuilder) Build(ctx context.Context, logger *zap.Logger, tracer logging.Tracer) (out *Dependencies, err error) {
	b.mustProvide(func() *zap.Logger { return logger })
	b.mustProvide(func() logging.Tracer { return tracer })
	b.mustProvide(func() context.Context { return ctx })
	b.mustProvide(newDataStore)

	logger.Info("building dependencies", zap.Reflect("builder", b))

	out = &Dependencies{
		dootaDepMissing: []string{},
	}

	if b.PGDSN != "" {
		err := b.dig.Invoke(func(dataStore datastore.Repository) {
			out.DataStore = dataStore
		})
		if err != nil {
			return nil, fmt.Errorf("failed to setup datastore: %w", err)
		}
		// out.DataStore, err = SetupDataStore(ctx, b.PGDSN, logger, tracer)
	} else {
		out.dootaDepMissing = append(out.dootaDepMissing, "datastore")
	}

	if b.KMSKeyPath != "" {
		out.AuthSigningKeyGetter, out.AuthTokenValidator, err = SetupKMS(ctx, b.KMSKeyPath, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to setup kms: %w", err)
		}
	} else {
		out.AuthSigningKeyGetter, out.AuthTokenValidator, err = SetupMockKMS(ctx, "", logger)
		if err != nil {
			return nil, fmt.Errorf("failed to setup kms: %w", err)
		}
	}

	if b.CorsURLRegexAllow != "" {
		urlRegex, err := regexp.Compile(b.CorsURLRegexAllow)
		if err != nil {
			return nil, fmt.Errorf("failed to compile CORS URL regex: %w", err)
		}
		out.CorsURLRegexAllow = urlRegex
	}

	if b.AIConfig != nil {
		debugStore, err := dstore.NewStore(b.AIConfig.OpenAIDebugLogsStore, "", "", false)
		if err != nil {
			return nil, fmt.Errorf("unable to create debug store: %w", err)
		}

		out.AIClient, err = ai.NewOpenAI(
			b.AIConfig.OpenAIKey,
			b.AIConfig.DefaultLLMModel,
			b.AIConfig.AdvanceLLMModel,
			ai.LangsmithConfig{
				ApiKey:      b.AIConfig.LangsmithApiKey,
				ProjectName: b.AIConfig.LangsmithProject,
			},
			debugStore,
			logger,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to create openai client: %w", err)
		}
	}

	if b.GoogleConfig != nil {
		logger.Info("setting up google",
			zap.Reflect("client_id", b.GoogleConfig.ClientID),
		)
		out.GoogleClient = google2.NewOauthClient(b.GoogleConfig.ClientID, b.GoogleConfig.ClientSecret, b.GoogleConfig.RedirectURL, logger)
	}

	if b.ConversationState != nil {
		out.ConversationState = state.NewCustomerCaseState(b.ConversationState.redisAddr, b.ConversationState.phoneCallStateTTL, logger, b.ConversationState.namespace, b.ConversationState.prefix)
	}

	return out, nil
}

type Dependencies struct {
	DataStore datastore.Repository

	AuthSigningKeyGetter crypto.SigningKeyGetter
	AuthTokenValidator   auth.TokenValidationFunc

	CorsURLRegexAllow *regexp.Regexp

	dootaDepMissing []string

	AIClient          *ai.Client
	GoogleClient      *google2.OauthClient
	ConversationState state.ConversationState
}
