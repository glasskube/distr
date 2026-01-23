package deploymenttargetlogs

import (
	"github.com/distr-sh/distr/api"
	"go.uber.org/zap/zapcore"
)

type Core struct {
	zapcore.LevelEnabler
	Collector Exporter
	Encoder   zapcore.Encoder
}

// Write implements [zapcore.Core].
func (pc *Core) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	buf, err := pc.Encoder.EncodeEntry(ent, fields)
	if err != nil {
		return err
	}
	return pc.Collector.ExportDeploymentTargetLogs(api.DeploymentTargetLogRecord{
		Timestamp: ent.Time,
		Severity:  ent.Level.String(),
		Body:      buf.String(),
	})
}

// Sync implements [zapcore.Core].
func (pc *Core) Sync() error {
	if s, ok := pc.Collector.(Syncer); ok {
		return s.Sync()
	}
	return nil
}

// Check implements [zapcore.Core].
func (pc *Core) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if pc.Enabled(ent.Level) {
		ce.AddCore(ent, pc)
	}
	return ce
}

// With implements [zapcore.Core].
func (pc *Core) With(fields []zapcore.Field) zapcore.Core {
	clone := pc.clone()
	for i := range fields {
		fields[i].AddTo(clone.Encoder)
	}
	return clone
}

func (pc *Core) clone() *Core {
	return &Core{
		LevelEnabler: pc.LevelEnabler,
		Collector:    pc.Collector,
		Encoder:      pc.Encoder.Clone(),
	}
}

var _ zapcore.Core = (*Core)(nil)
