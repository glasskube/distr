package svc

import (
	"context"

	"github.com/distr-sh/distr/internal/oidc"
	"go.uber.org/zap"
)

func (r *Registry) GetOIDCer() *oidc.OIDCer {
	return r.oidcer
}

func (r *Registry) createOIDCer(ctx context.Context, log *zap.Logger) (*oidc.OIDCer, error) {
	return oidc.NewOIDCer(ctx, log)
}
