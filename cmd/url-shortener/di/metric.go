package di

import (
	"context"

	"github.com/h3isenbug/url-shortener/internal/config"
	"github.com/h3isenbug/url-shortener/internal/monitoring"
)

func provideMetricCollector() (monitoring.MetricCollector, func()) {
	metricCollector := monitoring.NewPrometheusV1(config.Config.PrometheusMetricsPort)
	return metricCollector, func() {
		_ = metricCollector.Shutdown(context.Background())
	}
}
