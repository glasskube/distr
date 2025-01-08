package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/glasskube/cloud/internal/util"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	logger      = util.Require(zap.NewDevelopment())
	configFlags = genericclioptions.NewConfigFlags(true)
	interval    = 5 * time.Second
)

func init() {
	if intervalStr, ok := os.LookupEnv("GK_INTERVAL"); ok {
		interval = util.Require(time.ParseDuration(intervalStr))
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT)
		<-sigint
		logger.Info("received termination signal")
		cancel()
	}()
	tick := time.Tick(interval)
loop:
	for {
		select {
		case <-tick:
		case <-ctx.Done():
			break loop
		}

		var configuration action.Configuration
		err := configuration.Init(
			configFlags,
			"default",
			"secret",
			func(format string, v ...interface{}) { logger.Sugar().Debugf(format, v...) },
		)
		if err != nil {
			logger.Error("action config error", zap.Error(err))
			continue
		}
		upgradeAction := action.NewUpgrade(&configuration)
		upgradeAction.Install = true
		upgradeAction.Wait = true
		upgradeAction.Atomic = true
		upgradeAction.Namespace = "TODO: add namespace from deployment target"
		upgradeAction.RepoURL = "TODO: add repo url from application version"
		upgradeAction.Version = "TODO: add chart version from application version"

		values := map[string]any{} // TODO: add merged values from application version and deployment

		if chartPath, err := upgradeAction.LocateChart(
			"TODO: add chart name from application version", cli.New()); err != nil {
			panic(err)
		} else if chart, err := loader.Load(chartPath); err != nil {
			logger.Error("chart loading failed", zap.Error(err))
			continue
		} else if _, err = upgradeAction.RunWithContext(ctx,
			"TODO: add release name from deployment", chart, values); err != nil {
			logger.Error("helm upgrade failed", zap.Error(err))
		} else {
			logger.Info("helm upgrade successful")
		}
	}
	logger.Info("shutting down")
}
