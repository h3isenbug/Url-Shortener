package repository

import (
	"errors"
	"time"

	"github.com/h3isenbug/url-shortener/internal/monitoring"
)

var (
	ErrNotFound           = errors.New("not found")
	ErrUniquenessViolated = errors.New("requested operation violates uniqueness of a field")
)

type BaseMetricWrapper struct {
	MetricCollector monitoring.MetricCollector
	Name            string
}

func NewBaseMetricWrapper(metricCollector monitoring.MetricCollector, name string) *BaseMetricWrapper {
	return &BaseMetricWrapper{
		MetricCollector: metricCollector,
		Name:            name,
	}
}

func (w BaseMetricWrapper) RecordMetrics(method string, duration time.Duration, success bool) {
	var status = "OK"
	if !success {
		status = "FAILED"
	}

	w.MetricCollector.DependencyResponseTime(
		w.Name,
		method,
		status,
		duration,
	)
}
