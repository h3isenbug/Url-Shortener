package url

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	mathRand "math/rand"
	"strings"

	"github.com/h3isenbug/url-shortener/internal/repository"
	urlRepository "github.com/h3isenbug/url-shortener/internal/repository/url"
	"github.com/h3isenbug/url-shortener/internal/types"
	"github.com/h3isenbug/url-shortener/pkg/log"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type v1 struct {
	logger        log.Logger
	urlRepository urlRepository.Repository

	randomSlugLength int
}

func NewUrlServiceV1(logger log.Logger, urlRepository urlRepository.Repository, randomSlugLength int) Service {
	return &v1{
		logger:           logger,
		urlRepository:    urlRepository,
		randomSlugLength: randomSlugLength,
	}
}

func (s v1) GetOriginalUrl(ctx context.Context, slug string, newVisit bool) (originalUrl string, err error) {
	url, err := s.urlRepository.GetBySlug(ctx, slug)
	if err != nil {
		return "", fmt.Errorf("failed to get url by slug: %w", err)
	}

	if err := s.urlRepository.IncrementVisits(ctx, slug, newVisit); err != nil {
		return "", fmt.Errorf("failed to increment visit metrics: %w", err)
	}

	return url.OriginalUrl, nil
}

func (s v1) CreateShortUrl(ctx context.Context, originalUrl, recommendedShortLink string, accountID uint64) (string, error) {
	var shortLink string
	if recommendedShortLink != "" {
		shortLink = recommendedShortLink
	} else {
		shortLink = s.generateRandomString(s.randomSlugLength)
	}
	err := s.urlRepository.CreateShortUrl(ctx, originalUrl, shortLink, accountID)
	if errors.Is(err, repository.ErrUniquenessViolated) {
		if recommendedShortLink == "" {
			return "", fmt.Errorf("generated random string(%s) collided. this is very unlikely", shortLink)
		} else {
			return "", fmt.Errorf("recommended slug is taken: %w", err)
		}
	}
	if err != nil {
		return "", fmt.Errorf("failed to save short url: %w", err)
	}

	return shortLink, nil
}

func (s v1) GetAccountUrls(ctx context.Context, accountID uint64, cursor string) (items []types.Url, nextCursor string, err error) {
	return s.urlRepository.GetByAccountID(ctx, accountID, cursor)
}

func (s v1) SetUrlState(ctx context.Context, accountID uint64, slug string, disabled bool) error {
	if err := s.urlRepository.SetUrlState(ctx, accountID, slug, disabled); err != nil {
		return fmt.Errorf("failed to disable url(%s) of account(%d): %w", slug, accountID, err)
	}

	return nil
}

func (s v1) randomUint64() uint64 {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		s.logger.Error("failed to read from crypto/rand. using math/rand for now", map[string]interface{}{
			"errorMessage": err.Error(),
		})

		return mathRand.Uint64()
	}
	var generated uint64

	for i := 0; i < len(bytes); i++ {
		generated |= uint64(bytes[i] << i * 8)
	}

	return generated
}

func (s v1) generateRandomString(length int) string {
	var builder strings.Builder

	for i := 0; i < length; i++ {
		randUint := s.randomUint64()
		builder.WriteByte(
			charset[randUint%uint64(len(charset))],
		)
	}

	return builder.String()
}
