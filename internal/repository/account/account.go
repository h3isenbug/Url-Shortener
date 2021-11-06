package account

import (
	"context"
	"time"

	"github.com/h3isenbug/url-shortener/internal/monitoring"
	"github.com/h3isenbug/url-shortener/internal/repository"
	"github.com/h3isenbug/url-shortener/internal/types"
)

type Repository interface {
	Get(ctx context.Context, id uint64) (*types.Account, error)
	GetByEMail(ctx context.Context, email string) (*types.Account, error)
	Create(ctx context.Context, email, password string) error
}

type metricWrapper struct {
	*repository.BaseMetricWrapper

	wrapped Repository
}

func NewMetricWrapper(wrapped Repository, metricCollector monitoring.MetricCollector, name string) Repository {
	return &metricWrapper{
		BaseMetricWrapper: repository.NewBaseMetricWrapper(metricCollector, name),
		wrapped:           wrapped,
	}
}

func (w metricWrapper) Get(ctx context.Context, id uint64) (*types.Account, error) {
	startedAt := time.Now()
	account, err := w.wrapped.Get(ctx, id)
	w.RecordMetrics("Get", time.Now().Sub(startedAt), err == nil)

	return account, err
}

func (w metricWrapper) GetByEMail(ctx context.Context, email string) (*types.Account, error) {
	startedAt := time.Now()
	account, err := w.wrapped.GetByEMail(ctx, email)
	w.RecordMetrics("GetByEMail", time.Now().Sub(startedAt), err == nil)

	return account, err
}

func (w metricWrapper) Create(ctx context.Context, email, password string) error {
	startedAt := time.Now()
	err := w.wrapped.Create(ctx, email, password)
	w.RecordMetrics("Create", time.Now().Sub(startedAt), err == nil)

	return err
}
