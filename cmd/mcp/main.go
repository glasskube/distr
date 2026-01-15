package main

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"

	"github.com/distr-sh/distr/cmd/mcp/client"
	"github.com/distr-sh/distr/cmd/mcp/tools"
	"github.com/distr-sh/distr/internal/authkey"
	"github.com/distr-sh/distr/internal/buildconfig"
	"github.com/distr-sh/distr/internal/envutil"
	"github.com/distr-sh/distr/internal/util"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var log *zap.Logger

func authFromRequest(ctx context.Context, r *http.Request) context.Context {
	authHeader := r.Header.Get("Authorization")
	return client.WithAuthToken(ctx, authHeader)
}

func authFromEnv(ctx context.Context) context.Context {
	if token, err := envutil.RequireEnvParsedErr("DISTR_TOKEN", authkey.Parse); err == nil {
		return client.WithAuthToken(ctx, "AccessToken "+token.Serialize())
	}
	return ctx
}

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

	serveCmd.Flags().StringVarP(&serveOpts.transport, "transport", "t", "http", "transport type (stdio or http)")
}

func main() {
	err := func() error {
		defer func() { _ = log.Sync() }()

		err := rootCmd.Execute()
		if err != nil {
			log.Error("command returned error", zap.Error(err))
		}

		return err
	}()
	if err != nil {
		os.Exit(1)
	}
}

var serveCmd = &cobra.Command{
	Use: "serve",
	Run: serveCmdRun,
}

var serveOpts = struct {
	transport string
}{}

func serveCmdRun(cmd *cobra.Command, args []string) {
	mcpServer := server.NewMCPServer(
		"Distr",
		buildconfig.Version(),
		server.WithToolCapabilities(true),
	)

	var clientConfig *client.Config
	var err error

	switch serveOpts.transport {
	case "stdio":
		// For stdio mode, require DISTR_TOKEN from environment
		clientConfig, err = clientConfigFromEnv()
		if err != nil {
			log.Fatal("client config is invalid", zap.Error(err))
		}
		log.Sugar().Debugf("got client config: %v", clientConfig)

		c := client.NewClient(clientConfig)
		mgr := tools.NewManager(c)
		mgr.AddToolsToServer(mcpServer)

		log.Info("starting to serve in STDIO mode")
		err := server.ServeStdio(mcpServer, server.WithStdioContextFunc(authFromEnv))
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Fatal("serve failed", zap.Error(err))
		}
	case "http":
		// For HTTP mode, use token from request Authorization header
		clientConfig, err = clientConfigFromEnvWithoutToken()
		if err != nil {
			log.Fatal("client config is invalid", zap.Error(err))
		}

		c := client.NewClient(clientConfig)
		mgr := tools.NewManager(c)
		mgr.AddToolsToServer(mcpServer)

		log.Info("starting to serve in HTTP mode on :3001 and path /mcp")
		httpServer := server.NewStreamableHTTPServer(mcpServer, server.WithHTTPContextFunc(authFromRequest))
		if err := httpServer.Start(":3001"); err != nil {
			log.Fatal("serve failed", zap.Error(err))
		}
	default:
		log.Fatal("invalid transport type", zap.String("transport", serveOpts.transport))
	}
	log.Info("Distr MCP server shutting down")
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

func clientConfigFromEnvWithoutToken() (*client.Config, error) {
	opts := []client.ConfigOption{
		client.WithLogger(log),
		client.WithContextAuth(true),
	}

	if url, err := envutil.GetEnvParsedOrNilErr("DISTR_HOST", url.Parse); err != nil {
		return nil, err
	} else if url != nil {
		opts = append(opts, client.WithBaseURL(*url))
	}

	return client.NewConfig(opts...), nil
}
