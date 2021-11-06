package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type config struct {
	HTTPPort string `env:"HTTP_PORT"`

	DatabaseHost       string `env:"DB_HOST"`
	DatabasePort       int    `env:"DB_PORT"`
	DatabaseName       string `env:"DB_NAME"`
	DatabaseUser       string `env:"DB_USER"`
	DatabasePassword   string `env:"DB_PASSWORD"`
	DatabaseSSLEnabled bool   `env:"DB_SSL_ENABLED"`

	PrometheusMetricsPort int `env:"PROMETHEUS_METRICS_PORT"`

	SentryDSN string `env:"SENTRY_DSN"`

	RefreshTokenLength          int    `env:"REFRESH_TOKEN_LENGTH"`
	RefreshTokenLifespanSeconds int    `env:"REFRESH_TOKEN_LIFESPAN_SECONDS"`
	AccessTokenLifespanSeconds  int    `env:"ACCESS_TOKEN_LIFESPAN_SECONDS"`
	AccessTokenSecretFile       string `env:"ACCESS_TOKEN_SECRET_FILE"`
	AccessTokenCurrentKID       string `env:"ACCESS_TOKEN_CURRENT_KID"`

	RandomSlugLength int `env:"RANDOM_SLUG_LENGTH"`

	Hostname  string `env:"HOSTNAME"`
	DeployTag string `env:"DEPLOY_TAG"`

	ItemsPerPage int `env:"ITEMS_PER_PAGE"`

	GracefulShutdownPeriodSeconds int `env:"GRACEFUL_SHUTDOWN_PERIOD_SECONDS"`

	DashboardHost string `env:"DASHBOARD_HOST"`
	ShortUrlHost  string `env:"SHORT_URL_HOST"`

	RedisServer        string `env:"REDIS_SERVER"`
	RedisPassword      string `env:"REDIS_PASSWORD"`
	RedisDBForCache    int    `env:"REDIS_DB_FOR_CACHE"`
	UrlCacheTTLSeconds int    `env:"URL_CACHE_TTL_SECONDS"`
}

var Config config

func init() {
	t := reflect.TypeOf(Config)
	v := reflect.ValueOf(&Config).Elem()

	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("env")
		stringValue, found := os.LookupEnv(tag)
		if !found {
			_, _ = fmt.Fprintf(os.Stderr, "environment variable %s not set\n", tag)
			os.Exit(1)
		}

		switch t.Field(i).Type.Kind() {
		case reflect.String:
			v.Field(i).SetString(stringValue)
		case reflect.Slice:
			byteArrayValue, err := base64.StdEncoding.DecodeString(stringValue)
			if err != nil {
				panic(fmt.Sprintf("environment variable %s has incorrect value. expected base64.", tag))
			}
			v.Field(i).SetBytes(byteArrayValue)
		case reflect.Int:
			intValue, err := strconv.ParseInt(stringValue, 10, 32)
			if err != nil {
				panic(fmt.Sprintf("environment variable %s has incorrect value. expected int.", tag))
			}
			v.Field(i).SetInt(intValue)
		case reflect.Bool:
			v.Field(i).SetBool(strings.ToLower(stringValue) == "true")
		default:
			_, _ = fmt.Fprintf(os.Stderr, "unknown config field type: %s\n", t.Field(i).Type.Kind().String())
			os.Exit(1)
		}
	}
}
