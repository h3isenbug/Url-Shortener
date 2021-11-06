package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var defaultBuckets = []float64{.005, .02, .04, .07, .1, .15, .25, 0.5, 0.75, 1, 2, 3, 5, 10, 15, 20, 25, 30}

type prometheusV1 struct {
	httpResponseTimeMetric, dependencyResponseTimeMetric *prometheus.HistogramVec
	registry                                             *prometheus.Registry
	server                                               *http.Server
}

func NewPrometheusV1(port int) MetricCollector {
	p := &prometheusV1{
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			ReadTimeout:  time.Second * 2,
			WriteTimeout: time.Second * 5,
		},
		registry: prometheus.NewRegistry(),
		httpResponseTimeMetric: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_response_time_seconds",
			Help:    "response time of http endpoints in seconds",
			Buckets: defaultBuckets,
		}, []string{"method", "path", "status_code"}),
		dependencyResponseTimeMetric: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "dependency_response_time_seconds",
			Help:    "response time of dependencies in seconds",
			Buckets: defaultBuckets,
		}, []string{"name", "method", "status"}),
	}
	p.registry.MustRegister(p.httpResponseTimeMetric)
	p.registry.MustRegister(p.dependencyResponseTimeMetric)

	return p
}

func (p prometheusV1) HttpResponseTime(method, path string, statusCode int, duration time.Duration) {
	p.httpResponseTimeMetric.With(prometheus.Labels{
		"method": method, "path": path, "status_code": strconv.Itoa(statusCode),
	}).Observe(duration.Seconds())
}

func (p prometheusV1) DependencyResponseTime(name, method, status string, duration time.Duration) {
	p.dependencyResponseTimeMetric.With(prometheus.Labels{
		"method": method, "name": name, "status": status,
	}).Observe(duration.Seconds())
}

func (p prometheusV1) Start() error {
	handler := promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{})
	p.server.Handler = handler
	if err := p.server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
func (p prometheusV1) Shutdown(ctx context.Context) error {
	return p.server.Shutdown(ctx)
}
