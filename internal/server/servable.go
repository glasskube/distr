package server

import "context"

type Servable interface {
	Start(addr string) error
	Shutdown(ctx context.Context)
	WaitForShutdown()
}
