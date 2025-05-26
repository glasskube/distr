package jobs

import (
	"go.uber.org/zap"
)

type gocronLoggerAdapter struct {
	logger *zap.SugaredLogger
}

// Debug implements gocron.Logger.
func (g *gocronLoggerAdapter) Debug(msg string, args ...any) {
	g.logger.With(args...).Debugf(msg)
}

// Error implements gocron.Logger.
func (g *gocronLoggerAdapter) Error(msg string, args ...any) {
	g.logger.With(args...).Errorf(msg)
}

// Info implements gocron.Logger.
func (g *gocronLoggerAdapter) Info(msg string, args ...any) {
	g.logger.With(args...).Infof(msg)
}

// Warn implements gocron.Logger.
func (g *gocronLoggerAdapter) Warn(msg string, args ...any) {
	g.logger.With(args...).Warnf(msg)
}
