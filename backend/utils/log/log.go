package log

import (
	"bookstore/backend/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a production-ready Zap logger at the level specified in cfg.
// Falls back to "info" for any unrecognised level string.
func New(cfg config.LoggerConfig) *zap.Logger {
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	zapCfg := zap.NewProductionConfig()
	zapCfg.Level = zap.NewAtomicLevelAt(level)
	zapCfg.EncoderConfig.TimeKey = "ts"
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := zapCfg.Build(zap.AddCallerSkip(0))
	if err != nil {
		// If we cannot build the logger, panic early — the application
		// cannot operate safely without structured logging.
		panic("failed to initialise logger: " + err.Error())
	}

	return logger
}
