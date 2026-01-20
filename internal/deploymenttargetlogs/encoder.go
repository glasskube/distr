package deploymenttargetlogs

import (
	"github.com/mailru/easyjson/buffer"
	"go.uber.org/zap/zapcore"
)

type Encoder struct {
	zapcore.Encoder
}

func EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	return nil, nil
}
