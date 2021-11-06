package di

import (
	"github.com/h3isenbug/url-shortener/internal/config"
	"github.com/h3isenbug/url-shortener/pkg/log"
)

func provideLogger() (log.Logger, error) {
	return log.NewZapLoggingService(config.Config.SentryDSN)
}
