package url

import (
	"context"
	"time"

	"github.com/h3isenbug/url-shortener/internal/monitoring"
	"github.com/h3isenbug/url-shortener/internal/repository"
	"github.com/h3isenbug/url-shortener/internal/types"
)

type Repository interface {
	GetBySlug(ctx context.Context, slug string) (*types.Url, error)
	IncrementVisits(ctx context.Context, slug string, newVisit bool) error
	CreateShortUrl(ctx context.Context, originalUrl, slug string, accountID uint64) error
	GetByAccountID(ctx context.Context, accountID uint64, cursor string) (items []types.Url, nextCursor string, err error)
	SetUrlState(ctx context.Context, accountID uint64, slug string, disabled bool) error
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

func (w metricWrapper) GetBySlug(ctx context.Context, slug string) (*types.Url, error) {
	startedAt := time.Now()
	url, err := w.wrapped.GetBySlug(ctx, slug)
	w.RecordMetrics("GetBySlug", time.Now().Sub(startedAt), err == nil)

	return url, err
}

func (w metricWrapper) IncrementVisits(ctx context.Context, slug string, newVisit bool) error {
	startedAt := time.Now()
	err := w.wrapped.IncrementVisits(ctx, slug, newVisit)
	w.RecordMetrics("IncrementVisits", time.Now().Sub(startedAt), err == nil)

	return err
}

func (w metricWrapper) CreateShortUrl(ctx context.Context, originalUrl, slug string, accountID uint64) error {
	startedAt := time.Now()
	err := w.wrapped.CreateShortUrl(ctx, originalUrl, slug, accountID)
	w.RecordMetrics("CreateShortUrl", time.Now().Sub(startedAt), err == nil)

	return err
}

func (w metricWrapper) GetByAccountID(ctx context.Context, accountID uint64, cursor string) ([]types.Url, string, error) {
	startedAt := time.Now()
	items, nextCursor, err := w.wrapped.GetByAccountID(ctx, accountID, cursor)
	w.RecordMetrics("GetByAccountID", time.Now().Sub(startedAt), err == nil)

	return items, nextCursor, err
}

func (w metricWrapper) SetUrlState(ctx context.Context, accountID uint64, slug string, disabled bool) error {
	startedAt := time.Now()
	err := w.wrapped.SetUrlState(ctx, accountID, slug, disabled)
	w.RecordMetrics("SetUrlState", time.Now().Sub(startedAt), err == nil)

	return err
}
