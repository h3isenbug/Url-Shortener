package url

import (
	"context"
	"errors"

	"github.com/h3isenbug/url-shortener/internal/types"
)

var (
	ErrNotAuthorized = errors.New("user is not authorized to do the given action")
)

type Service interface {
	GetOriginalUrl(ctx context.Context, slug string, newVisit bool) (originalUrl string, err error)
	CreateShortUrl(ctx context.Context, originalUrl, recommendedSlug string, accountID uint64) (slug string, err error)
	GetAccountUrls(ctx context.Context, accountID uint64, cursor string) (items []types.Url, nextCursor string, err error)
	SetUrlState(ctx context.Context, accountID uint64, slug string, disabled bool) error
}
