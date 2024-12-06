package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type Server struct {
	server *http.Server
	logger *zap.Logger
}

func NewServer(handler http.Handler, logger *zap.Logger) *Server {
	server := &Server{
		server: &http.Server{
			Handler: handler,
		},
		logger: logger,
	}
	return server
}

func (s *Server) Start(addr string) error {
	s.server.Addr = addr
	s.logger.Sugar().Infof("starting listener on %v", s.server.Addr)
	if err := s.server.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
		return nil
	} else {
		return fmt.Errorf("could not start server: %w", err)
	}
}

func (s *Server) Shutdown(ctx context.Context) {
	s.logger.Warn("shutting down HTTP server")
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("error shutting down", zap.Error(err))
	}
}
