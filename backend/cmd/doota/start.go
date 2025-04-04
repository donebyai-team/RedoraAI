package main

import (
	"fmt"
	"github.com/shank318/doota/agents/vana"
	"github.com/shank318/doota/ai"
	"github.com/shank318/doota/app"
	"github.com/shank318/doota/auth"
	"github.com/shank318/doota/integrations"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"github.com/shank318/doota/portal"
	"github.com/shank318/doota/services"
	"os"
	"regexp"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/streamingfast/cli"
	"github.com/streamingfast/cli/sflags"
	tracing "github.com/streamingfast/sf-tracing"
	"golang.org/x/exp/maps"
)

var StartCmd = cli.Command(startCmdE,
	"start",
	"Starts the given applications, one of portal, extractor",
	cli.ArbitraryArgs(),
	cli.Flags(func(flags *pflag.FlagSet) {
		flags.Duration("common-phone-call-ttl", 5*time.Minute, cli.FlagDescription(`TTL to set in redis for a phone call`))
		flags.String("common-pubsub-project", "doota-local", "Google GCP Project")
		flags.String("common-gpt-model", "gpt-4o-2024-08-06", "GPT Model to use for message creator and categorization")
		flags.String("common-openai-api-key", "", "OpenAI API key")
		flags.String("common-openai-debug-store", "data/debugstore", "OpenAI debug store")
		flags.String("common-openai-organization", "", "OpenAI Organization")
		flags.String("common-langsmith-api-key", "", "Langsmith API key")
		flags.String("common-langsmith-project", "", "Langsmith project name")
		flags.Uint64("common-auto-mem-limit-percent", 0, "Automatically sets GOMEMLIMIT to a percentage of memory limit from cgroup (useful for container environments)")
		flags.Duration("spooler-db-polling-interval", 10*time.Second, "How often the spooler will check the database for new investigation")

		flags.String("portal-cors-url-regex-allow", "^.*", "Regex to allow CORS origin requests from, matched on the full URL (scheme, host, port, path, etc.), defaults to allow all")
		flags.String("portal-http-listen-addr", ":8787", "http listen address")

		flags.String("portal-fullstory-org-id", "", "FullStory org id")
		flags.String("portal-auth0-domain", "", "Auth0 tenant domain")
		flags.String("portal-auth0-portal-client-id", "", "Auth0 Portal AppFactory Client ID")
		flags.String("portal-auth0-portal-client-secret", "", "Auth0 Portal AppFactory Client Secret")
		flags.String("portal-auth0-api-redirect-uri", "http://localhost:8787/auth/callback", "The API Auth callback URL")
	}),
)

type App interface {
	cli.Shutter
	cli.RunnableContextError
}

type AppFactory func(cmd *cobra.Command, isAppReady func() bool) (App, error)

var appToFactory = map[string]AppFactory{
	"portal-api":   portalApp,
	"vana-spooler": vanaSpoolerApp,
}

func startCmdE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	main := cli.NewApplication(ctx)

	if len(args) == 0 {
		args = maps.Keys(appToFactory)
	}

	var apps []App
	for _, arg := range args {
		factory, found := appToFactory[arg]
		cli.Ensure(found, "Unknown app %q", arg)

		a, err := factory(cmd, main.IsReady)
		cli.NoError(err, "Unable to create app %q", arg)

		apps = append(apps, a)
	}

	err := setAutoMemoryLimit(sflags.MustGetUint64(cmd, "common-auto-mem-limit-percent"), zlog)
	if err != nil {
		return err
	}

	if os.Getenv("SF_TRACING") != "" {
		zlog.Info("setting up  tracing")
		if err := tracing.SetupOpenTelemetry(cmd.Context(), "loadlogic"); err != nil {
			return fmt.Errorf("failed to setup tracing: %w", err)
		}
	}

	for _, app := range apps {
		main.SuperviseAndStart(app)
	}

	shutdownUnreadyPeriod := sflags.MustGetDuration(cmd, "shutdown-unready-period")
	shutdownGracePeriod := sflags.MustGetDuration(cmd, "shutdown-grace-period")

	return main.WaitForTermination(zlog, shutdownUnreadyPeriod, shutdownGracePeriod)
}

func openAILangsmithLegacyHandling(cmd *cobra.Command, prefix string) (string, string, string, string, string) {
	openaiApiKey, openaiApiKeyLegacyFlagPresent := sflags.MustGetStringProvided(cmd, prefix+"-openai-api-key")
	openaiOrganization, openaiOrganizationLegacyFlagPresent := sflags.MustGetStringProvided(cmd, prefix+"-openai-organization")
	openaiDebugStore, openaiDebugStoreLegacyFlagPresent := sflags.MustGetStringProvided(cmd, prefix+"-openai-debug-store")
	langsmithApiKey, langsmithApiKeyLegacyFlagPresent := sflags.MustGetStringProvided(cmd, prefix+"-langsmith-api-key")
	langsmithProject, langsmithProjectLegacyFlagPresent := sflags.MustGetStringProvided(cmd, prefix+"-langsmith-project")

	if !openaiApiKeyLegacyFlagPresent {
		openaiApiKey = sflags.MustGetString(cmd, "common-openai-api-key")
	}

	if !openaiOrganizationLegacyFlagPresent {
		openaiOrganization = sflags.MustGetString(cmd, "common-openai-organization")
	}

	if !openaiDebugStoreLegacyFlagPresent {
		openaiDebugStore = sflags.MustGetString(cmd, "common-openai-debug-store")
	}

	if !langsmithApiKeyLegacyFlagPresent {
		langsmithApiKey = sflags.MustGetString(cmd, "common-langsmith-api-key")
	}

	if !langsmithProjectLegacyFlagPresent {
		langsmithProject = sflags.MustGetString(cmd, "common-langsmith-project")
	}

	return openaiApiKey, openaiOrganization, openaiDebugStore, langsmithApiKey, langsmithProject
}

func vanaSpoolerApp(cmd *cobra.Command, isAppReady func() bool) (App, error) {
	openaiApiKey, openaiOrganization, openaiDebugStore, langsmithApiKey, langsmithProject := openAILangsmithLegacyHandling(cmd, "common")
	deps, err := app.NewDependenciesBuilder().
		WithDataStore(sflags.MustGetString(cmd, "pg-dsn")).
		WithAI(
			openaiApiKey,
			openaiOrganization,
			openaiDebugStore,
			langsmithApiKey,
			langsmithProject,
		).
		WithConversationState(
			sflags.MustGetString(cmd, "redis-addr"),
			sflags.MustGetDuration(cmd, "common-phone-call-ttl"),
		).
		Build(cmd.Context(), zlog, tracer)
	if err != nil {
		return nil, err
	}

	logger := zlog.Named("spooler")

	gptModel, err := ai.ParseGPTModel(sflags.MustGetString(cmd, "common-gpt-model"))
	if err != nil {
		return nil, fmt.Errorf("initiated extractor with invalid gpt model: %w", err)
	}

	integrationsFactory := integrations.NewFactory(deps.DataStore, logger)
	caseInvestigator := vana.NewCaseInvestigator(gptModel, deps.DataStore, deps.AIClient, logger, deps.ConversationState)

	return vana.New(
		deps.DataStore,
		deps.AIClient,
		gptModel,
		deps.ConversationState,
		caseInvestigator,
		integrationsFactory,
		1000,
		10,
		sflags.MustGetDuration(cmd, "spooler-db-polling-interval"),
		isAppReady,
		logger,
	), nil
}

func portalApp(cmd *cobra.Command, isAppReady func() bool) (App, error) {
	openaiApiKey, openaiOrganization, openaiDebugStore, langsmithApiKey, langsmithProject := openAILangsmithLegacyHandling(cmd, "common")
	deps, err := app.NewDependenciesBuilder().
		WithDataStore(sflags.MustGetString(cmd, "pg-dsn")).
		//WithKMSKeyPath(sflags.MustGetString(cmd, "jwt-kms-keypath")).
		WithCORSURLRegexAllow(sflags.MustGetString(cmd, "portal-cors-url-regex-allow")).
		WithConversationState(
			sflags.MustGetString(cmd, "redis-addr"),
			sflags.MustGetDuration(cmd, "common-phone-call-ttl"),
		).
		WithAI(
			openaiApiKey,
			openaiOrganization,
			openaiDebugStore,
			langsmithApiKey,
			langsmithProject,
		).
		Build(cmd.Context(), zlog, tracer)
	if err != nil {
		return nil, err
	}

	whitelistDomains := []*regexp.Regexp{
		regexp.MustCompile(".*localhost"),
		regexp.MustCompile(".*127.0.0.1"),
		regexp.MustCompile(`.*\.donebyai.team`),
	}

	authenticator := auth.NewAuthenticator(deps.AuthTokenValidator, deps.DataStore, zlog)

	logger := zlog.Named("portal")

	gptModel, err := ai.ParseGPTModel(sflags.MustGetString(cmd, "common-gpt-model"))
	if err != nil {
		return nil, fmt.Errorf("initiated extractor with invalid gpt model: %w", err)
	}

	integrationsFactory := integrations.NewFactory(deps.DataStore, logger)

	caseInvestigator := vana.NewCaseInvestigator(gptModel, deps.DataStore, deps.AIClient, logger, deps.ConversationState)

	vanaWebhookHandler := vana.NewVanaWebhookHandler(
		deps.DataStore,
		deps.ConversationState,
		caseInvestigator,
		integrationsFactory,
		logger,
	)

	authConfig := &services.Auth0Config{
		Auth0PortalClientID:     sflags.MustGetString(cmd, "portal-auth0-portal-client-id"),
		Auth0PortalClientSecret: sflags.MustGetString(cmd, "portal-auth0-portal-client-secret"),
		Auth0ApiRedirectURL:     sflags.MustGetString(cmd, "portal-auth0-api-redirect-uri"),
		Auth0Domain:             sflags.MustGetString(cmd, "portal-auth0-domain"),
	}

	// TODO: Understand how to setup this as part of an auth use case
	config := &pbportal.Config{
		Auth0Domain:    authConfig.Auth0Domain,
		Auth0ClientId:  authConfig.Auth0PortalClientID,
		Auth0Scope:     "openid email",
		FullStoryOrgId: sflags.MustGetString(cmd, "portal-fullstory-org-id"),
	}

	authUsecase, err := services.NewAuthUsecase(cmd.Context(), authConfig, deps.DataStore, deps.AuthSigningKeyGetter, zlog)
	if err != nil {
		return nil, fmt.Errorf("unable to create auth usecase: %w", err)
	}

	p := portal.New(
		authenticator,
		services.NewCustomerCaseServiceImpl(deps.DataStore),
		authUsecase,
		services.NewCreateKeywordImpl(deps.DataStore),
		vanaWebhookHandler,
		deps.DataStore,
		sflags.MustGetString(cmd, "portal-http-listen-addr"),
		deps.CorsURLRegexAllow,
		config,
		whitelistDomains,
		isAppReady,
		zlog.Named("portal"),
		tracer,
	)
	return p, nil
}
