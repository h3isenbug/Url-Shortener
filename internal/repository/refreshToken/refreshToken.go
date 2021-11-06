package refreshToken

import (
	"context"
	"time"

	"github.com/h3isenbug/url-shortener/internal/monitoring"
	"github.com/h3isenbug/url-shortener/internal/repository"
	"github.com/h3isenbug/url-shortener/internal/types"
)

type Repository interface {
	Create(ctx context.Context, accountID uint64, token string, lifespan time.Duration) (*types.RefreshToken, error)
	CreateWithFamily(ctx context.Context, accountID uint64, token string, lifespan time.Duration, family uint64) (*types.RefreshToken, error)
	Get(ctx context.Context, token string) (*types.RefreshToken, error)
	Disable(ctx context.Context, id uint64) error
	SetCompromisedState(ctx context.Context, family uint64) error
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

func (w metricWrapper) Create(ctx context.Context, accountID uint64, token string, lifespan time.Duration) (*types.RefreshToken, error) {
	startedAt := time.Now()
	refreshToken, err := w.wrapped.Create(ctx, accountID, token, lifespan)
	w.RecordMetrics("Create", time.Now().Sub(startedAt), err == nil)

	return refreshToken, err
}

func (w metricWrapper) CreateWithFamily(ctx context.Context, accountID uint64, token string, lifespan time.Duration, family uint64) (*types.RefreshToken, error) {
	startedAt := time.Now()
	refreshToken, err := w.wrapped.CreateWithFamily(ctx, accountID, token, lifespan, family)
	w.RecordMetrics("CreateWithFamily", time.Now().Sub(startedAt), err == nil)

	return refreshToken, err
}

func (w metricWrapper) Get(ctx context.Context, token string) (*types.RefreshToken, error) {
	startedAt := time.Now()
	refreshToken, err := w.wrapped.Get(ctx, token)
	w.RecordMetrics("Get", time.Now().Sub(startedAt), err == nil)

	return refreshToken, err
}

func (w metricWrapper) Disable(ctx context.Context, id uint64) error {
	startedAt := time.Now()
	err := w.wrapped.Disable(ctx, id)
	w.RecordMetrics("Disable", time.Now().Sub(startedAt), err == nil)

	return err
}

func (w metricWrapper) SetCompromisedState(ctx context.Context, family uint64) error {
	startedAt := time.Now()
	err := w.wrapped.SetCompromisedState(ctx, family)
	w.RecordMetrics("SetCompromisedState", time.Now().Sub(startedAt), err == nil)

	return err
}
