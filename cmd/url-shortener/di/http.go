package di

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/h3isenbug/url-shortener/internal/config"
	"github.com/h3isenbug/url-shortener/internal/monitoring"
	presentation "github.com/h3isenbug/url-shortener/internal/presentation/http"
	"github.com/h3isenbug/url-shortener/internal/service/authentication"
	"github.com/h3isenbug/url-shortener/internal/service/url"
	"github.com/h3isenbug/url-shortener/pkg/log"
)

func provideHTTPServer(logger log.Logger, router *mux.Router) (*http.Server, func()) {
	server := &http.Server{
		Addr:    ":" + config.Config.HTTPPort,
		Handler: router,
	}
	return server, func() {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			time.Duration(config.Config.GracefulShutdownPeriodSeconds)*time.Second,
		)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Warn("error while shutting server down", map[string]interface{}{
				"errorMessage": err.Error(),
			})
		}
	}
}

func provideMuxRouter(
	logger log.Logger,
	authenticationService authentication.Service,
	authHandler presentation.AuthenticationAPI,
	urlHandler presentation.UrlAPI,
	metricCollector monitoring.MetricCollector,
) *mux.Router {
	router := mux.NewRouter()
	router.Use(presentation.GorillaMuxURLParamMiddleware)
	router.Use(presentation.GorillaHttpMetricsMiddleware(metricCollector))

	dashboardRouter := router.Host(config.Config.DashboardHost).PathPrefix("/api").Subrouter()

	urlRouter := dashboardRouter.PathPrefix("/url").Subrouter()
	urlRouter.Use(presentation.NewAuthMiddlewareV1(logger, authenticationService).Intercept)

	urlRouter.Methods("GET").HandlerFunc(urlHandler.GetMyUrls)
	urlRouter.Methods("POST").HandlerFunc(urlHandler.CreateShortUrl)
	urlRouter.Methods("PATCH").Path("/{slug:[0-9A-Za-z]+}").HandlerFunc(urlHandler.SetUrlState)
	urlRouter.Methods("GET").HandlerFunc(urlHandler.GetMyUrls)

	authRouter := dashboardRouter.PathPrefix("/auth").Subrouter()
	authRouter.Path("/login").Methods("POST").HandlerFunc(authHandler.Login)
	authRouter.Path("/register").Methods("POST").HandlerFunc(authHandler.Register)
	authRouter.Path("/renew").Methods("POST").HandlerFunc(authHandler.RenewAccessToken)

	router.Host(config.Config.ShortUrlHost).Methods("GET").Path("/{slug:[0-9A-Za-z]+}").HandlerFunc(urlHandler.GetOriginalUrl)

	return router
}

func provideAuthenticationAPI(logger log.Logger, authenticationService authentication.Service) presentation.AuthenticationAPI {
	return presentation.NewAuthenticationAPIV1(logger, authenticationService)
}

func provideUrlAPI(logger log.Logger, urlService url.Service) presentation.UrlAPI {
	return presentation.NewUrlAPIV1(logger, urlService)
}
