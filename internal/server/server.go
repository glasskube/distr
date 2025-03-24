package server

import "context"

type Server interface {
	Start(addr string) error
	Shutdown(ctx context.Context)
	WaitForShutdown()
}
