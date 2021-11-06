package url

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/h3isenbug/url-shortener/internal/types"
	"github.com/h3isenbug/url-shortener/pkg/log"
)

type redisCacheV1 struct {
	redis     *redis.Client
	ttl       time.Duration
	logger    log.Logger
	nextLayer Repository
}

func NewRedisCacheV1(logger log.Logger, redisClient *redis.Client, cacheTTL time.Duration, nextLayer Repository) Repository {
	return &redisCacheV1{
		redis:     redisClient,
		ttl:       cacheTTL,
		logger:    logger,
		nextLayer: nextLayer,
	}
}

func generateCacheKey(slug string) string {
	return fmt.Sprintf("url-%s", slug)
}

func (r redisCacheV1) GetBySlug(ctx context.Context, slug string) (*types.Url, error) {
	var url types.Url
	value, err := r.redis.Get(ctx, generateCacheKey(slug)).Result()
	if err == redis.Nil {
		return r.nextLayer.GetBySlug(ctx, slug)
	}
	if err != nil {
		r.logger.Warn("failed to fetch url cache entry", map[string]interface{}{
			"slug":         slug,
			"cacheKey":     generateCacheKey(slug),
			"errorMessage": err.Error(),
		})
		return r.nextLayer.GetBySlug(ctx, slug)
	}

	if err := url.FromString(value); err != nil {
		r.logger.Warn("invalid cache entry for url. it will be removed", map[string]interface{}{
			"slug":         slug,
			"cacheKey":     generateCacheKey(slug),
			"cacheValue":   value,
			"errorMessage": err.Error(),
		})
		if err := r.redis.Del(ctx, slug).Err(); err != nil {
			r.logger.Warn("failed to remove bogus cache entry from redis", map[string]interface{}{
				"slug":         slug,
				"cacheKey":     generateCacheKey(slug),
				"errorMessage": err.Error(),
			})
		}
	}

	return &url, nil
}

func (r redisCacheV1) IncrementVisits(ctx context.Context, slug string, newVisit bool) error {
	err := r.nextLayer.IncrementVisits(ctx, slug, newVisit)
	if err != nil {
		return err
	}

	if err := r.redis.Del(ctx, generateCacheKey(slug)).Err(); err != nil {
		r.logger.Warn("failed to invalidate cache entry", map[string]interface{}{
			"slug":         slug,
			"cacheKey":     generateCacheKey(slug),
			"errorMessage": err.Error(),
		})
	}

	return nil
}

func (r redisCacheV1) CreateShortUrl(ctx context.Context, originalUrl, slug string, accountID uint64) error {
	return r.nextLayer.CreateShortUrl(ctx, originalUrl, slug, accountID)
}

func (r redisCacheV1) GetByAccountID(ctx context.Context, accountID uint64, cursor string) (items []types.Url, nextCursor string, err error) {
	return r.nextLayer.GetByAccountID(ctx, accountID, cursor)
}

func (r redisCacheV1) SetUrlState(ctx context.Context, accountID uint64, slug string, disabled bool) error {
	err := r.nextLayer.SetUrlState(ctx, accountID, slug, disabled)
	if err != nil {
		return err
	}

	if err := r.redis.Del(ctx, generateCacheKey(slug)).Err(); err != nil {
		r.logger.Warn("failed to invalidate cache entry", map[string]interface{}{
			"slug":         slug,
			"cacheKey":     generateCacheKey(slug),
			"errorMessage": err.Error(),
		})
	}

	return nil
}
