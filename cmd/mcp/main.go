package main

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/glasskube/distr/cmd/mcp/client"
	"github.com/glasskube/distr/cmd/mcp/tools"
	"github.com/glasskube/distr/internal/buildconfig"
	"github.com/glasskube/distr/internal/envutil"
	"github.com/glasskube/distr/internal/util"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	defaultBaseUrl = util.Require(url.Parse("https://app.distr.sh/"))
	log            = util.Require(zap.NewDevelopment())
)

var rootCmd = &cobra.Command{
	Use:     "distr-mcp",
	Version: buildconfig.Version(),
}

func init() {
	rootCmd.AddCommand(serveCmd)
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

func serveCmdRun(cmd *cobra.Command, args []string) {
	clientConfig, err := clientConfigFromEnv()
	if err != nil {
		log.Fatal("client config is invalid", zap.Error(err))
	}

	mcpServer := server.NewMCPServer(
		"distr-mcp",
		buildconfig.Version(),
		server.WithResourceCapabilities(true, true),
		server.WithToolCapabilities(true),
		server.WithPromptCapabilities(true),
	)

	client := client.NewClient(clientConfig)
	mgr := tools.NewManager(client)

	mgr.AddToServer(mcpServer)

	if err := server.ServeStdio(mcpServer); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal("serve failed", zap.Error(err))
	}
	log.Info("distr-mcp shutting down")
}

func clientConfigFromEnv() (*client.Config, error) {
	if token, err := envutil.RequireEnvErr("DISTR_TOKEN"); err != nil {
		return nil, err
	} else if url, err := envutil.GetEnvParsedOrDefaultErr(
		"DISTR_HOST",
		url.Parse,
		defaultBaseUrl,
	); err != nil {
		return nil, err
	} else {
		return &client.Config{
			Token:      token,
			BaseUrl:    url,
			HttpClient: http.DefaultClient,
		}, nil
	}
}
