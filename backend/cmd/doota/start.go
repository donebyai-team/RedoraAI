package main

import (
	"fmt"
	"github.com/shank318/doota/app"
	"github.com/shank318/doota/auth"
	"github.com/shank318/doota/portal"
	"os"
	"regexp"

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
		flags.String("common-pubsub-project", "doota-local", "Google GCP Project")
		flags.String("common-gpt-model", "gpt-4o-2024-08-06", "GPT Model to use for message creator and categorization")
		flags.String("common-openai-api-key", "", "OpenAI API key")
		flags.String("common-openai-debug-store", "gs://doota-main-ai-debug-dev", "OpenAI debug store")
		flags.String("common-openai-organization", "", "OpenAI Organization")
		flags.String("common-langsmith-api-key", "", "Langsmith API key")
		flags.String("common-langsmith-project", "", "Langsmith project name")

		flags.String("portal-http-listen-addr", ":8787", "http listen address")
		flags.String("portal-auth0-domain", "", "Auth0 tenant domain")
		flags.String("portal-auth0-portal-client-id", "", "Auth0 Portal AppFactory Client ID")
		flags.String("portal-auth0-portal-client-secret", "", "Auth0 Portal AppFactory Client Secret")
		flags.String("portal-auth0-api-redirect-uri", "http://localhost:8787/auth/callback", "The API Auth callback URL")
		flags.String("portal-fullstory-org-id", "", "FullStory org id")
		flags.String("quote-pubsub-partial-quotes-subscription", "quote-partial-quotes-dev", "Pubsub partial quote quote service subscription")
	}),
)

type App interface {
	cli.Shutter
	cli.RunnableContextError
}

type AppFactory func(cmd *cobra.Command, isAppReady func() bool) (App, error)

var appToFactory = map[string]AppFactory{
	"portal-api": portalApp,
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

func portalApp(cmd *cobra.Command, isAppReady func() bool) (App, error) {

	deps, err := app.NewDependenciesBuilder().
		WithDataStore(sflags.MustGetString(cmd, "pg-dsn")).
		WithKMSKeyPath(sflags.MustGetString(cmd, "jwt-kms-keypath")).
		WithCORSURLRegexAllow(sflags.MustGetString(cmd, "portal-cors-url-regex-allow")).
		Build(cmd.Context(), zlog, tracer)
	if err != nil {
		return nil, err
	}

	whitelistDomains := []*regexp.Regexp{
		regexp.MustCompile(".*localhost"),
		regexp.MustCompile(".*127.0.0.1"),
		regexp.MustCompile(`.*\.dootaai.com`),
	}

	authenticator := auth.NewAuthenticator(deps.AuthTokenValidator, deps.DataStore, zlog)

	p := portal.New(
		authenticator,
		deps.DataStore,
		sflags.MustGetString(cmd, "portal-http-listen-addr"),
		deps.CorsURLRegexAllow,
		whitelistDomains,
		isAppReady,
		zlog.Named("portal"),
		tracer,
	)
	return p, nil
}
