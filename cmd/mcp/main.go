package main

import (
	"context"
	"errors"
	"net/url"

	"github.com/glasskube/distr/cmd/mcp/client"
	"github.com/glasskube/distr/cmd/mcp/tools"
	"github.com/glasskube/distr/internal/authkey"
	"github.com/glasskube/distr/internal/buildconfig"
	"github.com/glasskube/distr/internal/envutil"
	"github.com/glasskube/distr/internal/util"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var log *zap.Logger

var rootCmd = &cobra.Command{
	Use:     "distr-mcp",
	Version: buildconfig.Version(),
}

func init() {
	if buildconfig.IsRelease() {
		log = util.Require(zap.NewProduction())
	} else {
		log = util.Require(zap.NewDevelopment())
	}

	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().BoolVar(&serveOpts.sse, "sse", false, "start server with SSE (otherwise STDIO is used)")
}

func main() {
	defer func() { _ = log.Sync() }()
	if err := rootCmd.Execute(); err != nil {
		log.Fatal("command returned error", zap.Error(err))
	}
}

var serveCmd = &cobra.Command{
	Use: "serve",
	Run: serveCmdRun,
}

var serveOpts = struct {
	sse bool
}{}

func serveCmdRun(cmd *cobra.Command, args []string) {
	clientConfig, err := clientConfigFromEnv()
	if err != nil {
		log.Fatal("client config is invalid", zap.Error(err))
	}

	log.Sugar().Debugf("got client config: %v", clientConfig)

	mcpServer := server.NewMCPServer(
		"distr-mcp",
		buildconfig.Version(),
		server.WithResourceCapabilities(true, true),
		server.WithToolCapabilities(true),
		server.WithPromptCapabilities(true),
	)

	client := client.NewClient(clientConfig)
	mgr := tools.NewManager(client)

	mgr.AddToolsToServer(mcpServer)

	if serveOpts.sse {
		log.Info("starting to serve in SSE mode")
		if err := server.NewSSEServer(mcpServer).Start(":3001"); err != nil {
			log.Fatal("serve failed", zap.Error(err))
		}
	} else {
		log.Info("starting to serve in STDIO mode")
		if err := server.ServeStdio(mcpServer); err != nil && !errors.Is(err, context.Canceled) {
			log.Fatal("serve failed", zap.Error(err))
		}
	}
	log.Info("distr-mcp shutting down")
}

func clientConfigFromEnv() (*client.Config, error) {
	opts := []client.ConfigOption{
		client.WithLogger(log),
	}

	if token, err := envutil.RequireEnvParsedErr("DISTR_TOKEN", authkey.Parse); err != nil {
		return nil, err
	} else {
		opts = append(opts, client.WithToken(token))
	}

	if url, err := envutil.GetEnvParsedOrNilErr("DISTR_HOST", url.Parse); err != nil {
		return nil, err
	} else if url != nil {
		opts = append(opts, client.WithBaseURL(*url))
	}

	return client.NewConfig(opts...), nil
}
