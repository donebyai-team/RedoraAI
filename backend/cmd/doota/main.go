package main

import (
	_ "net/http/pprof"
	"time"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	. "github.com/streamingfast/cli"
	"github.com/streamingfast/logging"
)

// Version value, injected via go build `ldflags` at build time
var version = "dev"
var zlog, tracer = logging.RootLogger("doota", "github.com/shank318/doota/cmd/doota")

func init() {
	logging.InstantiateLoggers()
}

func main() {
	Run("doota", "Doota Management & Backend CLI",

		StartCmd,
		ToolsGroup,
		ConfigureViper("DOOTA"),
		ConfigureVersion(version),

		PersistentFlags(
			func(flags *pflag.FlagSet) {
				flags.Duration("delay-before-start", 0, "[OPERATOR] Amount of time to wait before starting any internal processes, can be used to perform to maintenance on the pod before actually letting it starts")
				flags.String("log-format", "text", "Format for logging to stdout. Either 'text' or 'stackdriver'. When 'text', if the standard output is detected to be interactive, colored text is output, otherwise non-colored text.")
				flags.String("metrics-listen-addr", "localhost:9102", "[OPERATOR] If non-empty, the process will listen on this address for Prometheus metrics request(s)")
				flags.String("pprof-listen-addr", "localhost:6060", "[OPERATOR] If non-empty, the process will listen on this address for pprof analysis (see https://golang.org/pkg/net/http/pprof/)")
				flags.Duration("shutdown-unready-period", 0*time.Second, "[OPERATOR] If non-zero, the process upon receiving the first Ctrl-C will be marked unready for this period of time before being shutdown allowing orchestrators to drain connections and remove the pod from the load-balancer")
				flags.Duration("shutdown-grace-period", 5*time.Second, "[OPERATOR] If non-zero, the process upon receiving the first Ctrl-C and after the elapsed unready period (if set) will give this period of time to components shutdown gracefully before being forced killed")
				flags.String("gcp-project", "doota-main", "GCP project name")
				flags.String("pg-dsn", "postgresql://dev-node:insecure-change-me-in-prod@localhost:5432/dev-node?enable_incremental_sort=off&sslmode=disable", "PostgreSQL DSN, set to empty to disable")
			},
		),
		AfterAllHook(func(cmd *cobra.Command) {
			cmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
				if err := setupCmd(cmd); err != nil {
					return err
				}
				return nil
			}
		}),
	)
}
