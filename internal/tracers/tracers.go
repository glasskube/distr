package tracers

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/multierr"
)

type Tracers struct {
	DefaultProvider  trace.TracerProvider
	AlwaysProvider   trace.TracerProvider
	AgentProvider    trace.TracerProvider
	RegistryProvider trace.TracerProvider
}

func (t *Tracers) Default() trace.TracerProvider {
	return otel.GetTracerProvider()
}

func (t *Tracers) Always() trace.TracerProvider {
	return t.AlwaysProvider
}

func (t *Tracers) Agent() trace.TracerProvider {
	if t.AgentProvider != nil {
		return t.AgentProvider
	}
	return t.Default()
}

func (t *Tracers) Registry() trace.TracerProvider {
	if t.RegistryProvider != nil {
		return t.RegistryProvider
	}
	return t.Default()
}

func (t *Tracers) Shutdown(ctx context.Context) error {
	return multierr.Combine(
		shutdown(ctx, t.DefaultProvider),
		shutdown(ctx, t.AlwaysProvider),
		shutdown(ctx, t.AgentProvider),
		shutdown(ctx, t.RegistryProvider),
	)
}

func shutdown(ctx context.Context, t trace.TracerProvider) error {
	type shutdown interface {
		Shutdown(ctx context.Context) error
	}
	if t != nil {
		if s, ok := t.(shutdown); ok {
			return s.Shutdown(ctx)
		}
	}
	return nil
}
