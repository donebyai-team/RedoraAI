package app

import (
	"context"
	"fmt"
	"github.com/shank318/doota/agents/state"
	"github.com/shank318/doota/ai"
	"github.com/shank318/doota/auth"
	"github.com/shank318/doota/auth/crypto"
	"github.com/shank318/doota/datastore"
	"github.com/streamingfast/dstore"
	"regexp"
	"time"

	"github.com/streamingfast/logging"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

type DependenciesBuilder struct {
	PGDSN              string
	KMSKeyPath         string
	CorsURLRegexAllow  string
	AttachmentStoreURL string
	PubsubGCPProject   string
	Processor          bool
	AIConfig           *AIConfig
	ConversationState  *conversationState

	dig *dig.Container
}

func NewDependenciesBuilder() *DependenciesBuilder {
	return &DependenciesBuilder{
		dig: dig.New(),
	}
}

type conversationState struct {
	redisAddr                  string
	investigationTTL           time.Duration
	investigationRetryCooldown time.Duration
}

type AIConfig struct {
	OpenAIKey            string
	OpenAIOrganization   string
	OpenAIDebugLogsStore string
	LangsmithApiKey      string
	LangsmithProject     string
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

func (b *DependenciesBuilder) WithConversationState(redisAddr string, investigationHeartbeat time.Duration, investigationRetryCooldown time.Duration) *DependenciesBuilder {
	b.ConversationState = &conversationState{redisAddr, investigationHeartbeat * 2, investigationRetryCooldown}
	return b
}

func (b *DependenciesBuilder) WithAI(openAIKey string, openAIOrganization string, openAIDebugLogsStore string, langsmithApiKey string, langsmithProject string) *DependenciesBuilder {
	b.AIConfig = &AIConfig{
		OpenAIKey:            openAIKey,
		OpenAIOrganization:   openAIOrganization,
		OpenAIDebugLogsStore: openAIDebugLogsStore,
		LangsmithApiKey:      langsmithApiKey,
		LangsmithProject:     langsmithProject,
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
			b.AIConfig.OpenAIOrganization,

			ai.LangsmithConfig{
				ApiKey:      b.AIConfig.LangsmithApiKey,
				ProjectName: b.AIConfig.LangsmithProject,
			},
			debugStore,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to create openai client: %w", err)
		}
	}

	if b.ConversationState != nil {
		out.ConversationState = state.NewCustomerCaseState(b.ConversationState.redisAddr, b.ConversationState.investigationTTL, b.ConversationState.investigationRetryCooldown, logger)
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
	ConversationState state.ConversationState
}
