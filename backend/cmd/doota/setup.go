package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/KimMachineGun/automemlimit/memlimit"
	"github.com/shank318/doota/metrics"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/dmetrics"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
)

func setupCmd(cmd *cobra.Command) error {
	cmd.SilenceUsage = true

	if rootWorkingDirectory := viper.GetString("global-root"); rootWorkingDirectory != "" {
		zlog.Info("changing working directory", zap.String("directory", rootWorkingDirectory))
		if err := os.Chdir(rootWorkingDirectory); err != nil {
			return err
		}
	}

	delay := viper.GetDuration("global-delay-before-start")
	if delay > 0 {
		zlog.Info("sleeping to respect delay before start setting", zap.Duration("delay", delay))
		time.Sleep(delay)
	}

	if v := viper.GetString("global-metrics-listen-addr"); v != "" {
		zlog.Info("starting prometheus metrics server", zap.String("listen_addr", v))

		dmetrics.Register(metrics.MetricSet)

		go dmetrics.Serve(v)
	}

	if v := viper.GetString("global-pprof-listen-addr"); v != "" {
		go func() {
			zlog.Info("starting pprof server", zap.String("listen_addr", v))
			err := http.ListenAndServe(v, nil)
			if err != nil {
				zlog.Debug("unable to start profiling server", zap.Error(err), zap.String("listen_addr", v))
			}
		}()
	}
	setupLogger(viper.GetString("global-log-format"))
	return nil
}

func setupLogger(logFormat string) {
	options := []logging.InstantiateOption{
		logging.WithSwitcherServerAutoStart(),
		logging.WithDefaultLevel(zap.InfoLevel),
		logging.WithConsoleToStdout(),
	}
	options = append(options, logging.WithConsoleToStderr())
	if logFormat == "stackdriver" || logFormat == "json" {
		options = append(options, logging.WithProductionLogger())
	}

	logging.InstantiateLoggers(options...)
}

func setAutoMemoryLimit(limit uint64, logger *zap.Logger) error {
	if limit != 0 {
		if limit > 100 {
			return fmt.Errorf("cannot set common-auto-mem-limit-percent above 100")
		}
		logger.Info("setting GOMEMLIMIT relative to available memory", zap.Uint64("percent", limit))
		memlimit.SetGoMemLimit(float64(limit) / 100)
	}
	return nil
}
