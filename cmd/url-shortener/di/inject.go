//+build wireinject

package di

import (
	"net/http"
	"sync"

	"github.com/google/wire"
	"github.com/h3isenbug/url-shortener/internal/monitoring"
	"github.com/h3isenbug/url-shortener/pkg/log"
)

type App struct {
	httpServer      *http.Server
	metricCollector monitoring.MetricCollector
	logger          log.Logger
}

func (a App) Start() {
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Warn("http server terminated", map[string]interface{}{
				"errorMessage": err,
			})
		}
	}()

	go func() {
		defer wg.Done()
		if err := a.metricCollector.Start(); err != nil && err != http.ErrServerClosed {
			a.logger.Warn("metric collector terminated", map[string]interface{}{
				"errorMessage": err,
			})
		}
	}()

	wg.Wait()
}
func (a App) Handler() http.Handler {
	return a.httpServer.Handler
}
func Inject() (*App, func(), error) {
	wire.Build(
		provideApp,

		provideMetricCollector,

		provideHTTPServer, provideMuxRouter,
		provideAuthenticationAPI,
		provideUrlAPI,

		provideAuthenticationService, provideAccessTokenSecrets,
		provideUrlService,

		provideLogger,

		provideSQLXConnection,
		provideAccountRepository, provideRefreshTokenRepository,
		provideUrlRepository,

		provideRedisClient,
	)
	return &App{}, func() {

	}, nil
}

func provideApp(logger log.Logger, httpServer *http.Server, metricCollector monitoring.MetricCollector) *App {
	return &App{
		httpServer:      httpServer,
		metricCollector: metricCollector,
		logger:          logger,
	}
}
