package log

import (
	"github.com/TheZeroSlave/zapsentry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogService struct {
	logger *zap.Logger
}

func NewZapLoggingService(sentryDSN string) (Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	logger = logger.WithOptions(zap.AddCallerSkip(1))

	if sentryDSN == "" {
		return &zapLogService{logger: logger}, nil
	}

	logger = logger.With(zapsentry.NewScope())

	core, err := zapsentry.NewCore(
		zapsentry.Configuration{
			Level:             zapcore.WarnLevel,
			EnableBreadcrumbs: true,
			BreadcrumbLevel:   zapcore.DebugLevel,
		},
		zapsentry.NewSentryClientFromDSN(sentryDSN),
	)
	if err != nil {
		return nil, err
	}

	return &zapLogService{
		logger: zapsentry.AttachCoreToLogger(core, logger),
	}, nil
}

func NewZapLoggingServiceWithSentry(sentryDSN string) (Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	logger.With(zapsentry.NewScope())

	cfg := zapsentry.Configuration{
		Level:             zapcore.WarnLevel,
		EnableBreadcrumbs: true,
		BreadcrumbLevel:   zapcore.DebugLevel,
	}

	core, err := zapsentry.NewCore(cfg, zapsentry.NewSentryClientFromDSN(sentryDSN))
	if err != nil {
		return nil, err
	}

	logger = zapsentry.AttachCoreToLogger(core, logger)

	return &zapLogService{logger: logger}, nil
}

func extrasToZapFields(extras []map[string]interface{}) []zap.Field {
	var fields []zap.Field
	for _, extra := range extras {
		for key, value := range extra {
			fields = append(fields, zap.Any(key, value))
		}
	}

	return fields
}
func (l zapLogService) Debug(msg string, extras ...map[string]interface{}) {
	l.logger.Debug(msg, extrasToZapFields(extras)...)
}

func (l zapLogService) Info(msg string, extras ...map[string]interface{}) {
	l.logger.Info(msg, extrasToZapFields(extras)...)
}

func (l zapLogService) Warn(msg string, extras ...map[string]interface{}) {
	l.logger.Warn(msg, extrasToZapFields(extras)...)
}

func (l zapLogService) Error(msg string, extras ...map[string]interface{}) {
	l.logger.Error(msg, extrasToZapFields(extras)...)
}

func (l zapLogService) Panic(msg string, extras ...map[string]interface{}) {
	l.logger.Panic(msg, extrasToZapFields(extras)...)
}
