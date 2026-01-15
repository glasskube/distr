package svc

import (
	"github.com/distr-sh/distr/internal/buildconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (r *Registry) GetLogger() *zap.Logger {
	return r.logger
}

func createLogger() *zap.Logger {
	if buildconfig.IsRelease() {
		config := zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		return zap.Must(config.Build())
	} else {
		return zap.Must(zap.NewDevelopment())
	}
}
