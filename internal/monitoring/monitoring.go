package monitoring

import (
	"context"
	"time"
)

type MetricCollector interface {
	HttpResponseTime(method, path string, statusCode int, duration time.Duration)
	DependencyResponseTime(name, method, status string, duration time.Duration)

	Start() error
	Shutdown(ctx context.Context) error
}
