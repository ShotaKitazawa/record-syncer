package logger

import (
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(isDebug bool) (logr.Logger, error) {
	// setup encorder
	enc := zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig())
	// setup syncer
	sink := zapcore.AddSync(os.Stdout)
	// setup log-level
	var level zap.AtomicLevel
	if isDebug {
		level = zap.NewAtomicLevelAt(zapcore.Level(-1)) // DEBUG
	} else {
		level = zap.NewAtomicLevelAt(zapcore.Level(0)) // INFO
	}
	// new
	logger := zap.New(zapcore.NewCore(enc, sink, level))
	return zapr.NewLogger(logger), nil
}
