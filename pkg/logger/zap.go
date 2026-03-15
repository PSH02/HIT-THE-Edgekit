package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	sugar *zap.SugaredLogger
}

func New(level string) Logger {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	switch level {
	case "debug":
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		cfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		cfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	l, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	return &zapLogger{sugar: l.Sugar()}
}

func (z *zapLogger) Debug(msg string, keysAndValues ...interface{}) {
	z.sugar.Debugw(msg, keysAndValues...)
}

func (z *zapLogger) Info(msg string, keysAndValues ...interface{}) {
	z.sugar.Infow(msg, keysAndValues...)
}

func (z *zapLogger) Warn(msg string, keysAndValues ...interface{}) {
	z.sugar.Warnw(msg, keysAndValues...)
}

func (z *zapLogger) Error(msg string, keysAndValues ...interface{}) {
	z.sugar.Errorw(msg, keysAndValues...)
}

func (z *zapLogger) With(keysAndValues ...interface{}) Logger {
	return &zapLogger{sugar: z.sugar.With(keysAndValues...)}
}

func (z *zapLogger) WithContext(ctx context.Context) Logger {
	return z
}
