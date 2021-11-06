package di

import (
	"github.com/h3isenbug/url-shortener/internal/config"
	urlRepository "github.com/h3isenbug/url-shortener/internal/repository/url"
	"github.com/h3isenbug/url-shortener/internal/service/url"
	"github.com/h3isenbug/url-shortener/pkg/log"
)

func provideUrlService(
	logger log.Logger,
	urlRepository urlRepository.Repository,
) url.Service {
	return url.NewUrlServiceV1(
		logger,
		urlRepository,
		config.Config.RandomSlugLength,
	)
}
