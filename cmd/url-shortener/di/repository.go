package di

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/h3isenbug/url-shortener/internal/config"
	"github.com/h3isenbug/url-shortener/internal/monitoring"
	"github.com/h3isenbug/url-shortener/internal/repository/account"
	"github.com/h3isenbug/url-shortener/internal/repository/refreshToken"
	"github.com/h3isenbug/url-shortener/internal/repository/url"
	"github.com/h3isenbug/url-shortener/internal/types"
	"github.com/h3isenbug/url-shortener/pkg/log"
	"github.com/jmoiron/sqlx"
)

func provideSQLXConnection(logger log.Logger) (*sqlx.DB, func(), error) {
	sslMode := "disable"
	if config.Config.DatabaseSSLEnabled {
		sslMode = "require"
	}

	con, err := sqlx.Connect(
		"postgres",
		fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?application_name=%s-%s-%s&sslmode=%s",
			config.Config.DatabaseUser, config.Config.DatabasePassword, config.Config.DatabaseHost,
			config.Config.DatabasePort, config.Config.DatabaseName, types.ServiceName, config.Config.DeployTag,
			config.Config.Hostname,
			sslMode,
		),
	)
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to connect to db: %w", err)
	}

	return con, func() {
		if err := con.Close(); err != nil {
			logger.Warn(
				"failed to close connection to db",
				map[string]interface{}{"errorMessage": err.Error()},
			)
		}
	}, nil
}

func provideRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     config.Config.RedisServer,
		Password: config.Config.RedisPassword,
		DB:       config.Config.RedisDBForCache,
	})
}

func provideAccountRepository(connection *sqlx.DB, metricCollector monitoring.MetricCollector) account.Repository {
	return account.NewMetricWrapper(
		account.NewPostgresRepositoryV1(connection),
		metricCollector,
		"AccountRepositoryPostgres",
	)
}

func provideRefreshTokenRepository(connection *sqlx.DB, metricCollector monitoring.MetricCollector) refreshToken.Repository {
	return refreshToken.NewMetricWrapper(
		refreshToken.NewPostgresRepositoryV1(connection),
		metricCollector,
		"RefreshTokenRepositoryPostgres",
	)
}

func provideUrlRepository(
	logger log.Logger, connection *sqlx.DB, redisClient *redis.Client,
	metricCollector monitoring.MetricCollector,
) url.Repository {

	dbLayer := url.NewMetricWrapper(
		url.NewPostgresRepositoryV1(connection, config.Config.ItemsPerPage),
		metricCollector,
		"UrlRepositoryPostgres",
	)

	return url.NewMetricWrapper(
		url.NewRedisCacheV1(
			logger, redisClient,
			time.Second*time.Duration(config.Config.UrlCacheTTLSeconds),
			dbLayer,
		),
		metricCollector,
		"UrlRepositoryRedis",
	)
}
