package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/distr-sh/distr/internal/env"
	"go.uber.org/zap"
)

type server struct {
	server           *http.Server
	logger           *zap.Logger
	shutdownComplete chan struct{}
}

func NewServer(handler http.Handler, logger *zap.Logger) Server {
	server := &server{
		server: &http.Server{
			Handler: handler,
		},
		logger:           logger,
		shutdownComplete: make(chan struct{}),
	}
	return server
}

func (s *server) Start(addr string) error {
	s.server.Addr = addr
	s.logger.Sugar().Infof("starting listener on %v", s.server.Addr)
	if err := s.server.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
		return nil
	} else {
		return fmt.Errorf("could not start server: %w", err)
	}
}

func (s *server) Shutdown(ctx context.Context) {
	if d := env.ServerShutdownDelayDuration(); d != nil {
		s.logger.Sugar().Warnf("shutting down HTTP server in %v", d)
		time.Sleep(*d)
	}
	s.logger.Warn("shutting down HTTP server")
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("error shutting down", zap.Error(err))
	}
	close(s.shutdownComplete)
}

func (s *server) WaitForShutdown() {
	<-s.shutdownComplete
	s.logger.Info("server shutdown complete")
}
