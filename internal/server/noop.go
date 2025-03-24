package server

import "context"

type noopServer struct{}

func (s *noopServer) Start(addr string) error {
	return nil
}

func (s *noopServer) Shutdown(ctx context.Context) {
}

func (s *noopServer) WaitForShutdown() {
}

func NewNoop() Servable {
	return &noopServer{}
}
