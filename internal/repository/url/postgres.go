package url

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/h3isenbug/url-shortener/internal/repository"
	"github.com/h3isenbug/url-shortener/internal/types"
	"github.com/jmoiron/sqlx"
)

type postgresV1 struct {
	con          *sqlx.DB
	itemsPerPage int
}

func NewPostgresRepositoryV1(connection *sqlx.DB, itemsPerPage int) Repository {
	return &postgresV1{
		con:          connection,
		itemsPerPage: itemsPerPage,
	}
}
func (r postgresV1) GetBySlug(ctx context.Context, slug string) (*types.Url, error) {
	var url types.Url
	err := r.con.GetContext(
		ctx, &url,
		"SELECT id, original_url, slug, total_visits, unique_visits, account_id, disabled, created_at FROM urls WHERE slug=$1",
		slug,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: url not found(by slug)", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch url: %w", err)
	}

	return &url, nil
}

func (r postgresV1) IncrementVisits(ctx context.Context, slug string, newVisit bool) error {
	var query = "UPDATE urls SET total_visits=total_visits+1 WHERE slug=$1"
	if newVisit {
		query = "UPDATE urls SET total_visits=total_visits+1, unique_visits=unique_visits+1 WHERE slug=$1"
	}

	result, err := r.con.ExecContext(ctx, query, slug)
	if err != nil {
		return fmt.Errorf("failed to update url visit metrics: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r postgresV1) CreateShortUrl(ctx context.Context, originalUrl, slug string, accountID uint64) error {
	_, err := r.con.ExecContext(
		ctx,
		"INSERT INTO urls(original_url, slug, account_id) VALUES ($1, $2, $3)",
		originalUrl, slug, accountID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert url: %w", err)
	}

	return nil
}

func (r postgresV1) GetByAccountID(ctx context.Context, accountID uint64, cursor string) ([]types.Url, string, error) {
	var urls []types.Url
	offset, _ := strconv.Atoi(cursor)
	err := r.con.SelectContext(
		ctx, &urls,
		`SELECT
       				id, original_url, slug, total_visits, unique_visits, account_id, disabled, created_at
			   FROM urls WHERE account_id=$1 ORDER BY created_at DESC OFFSET $2 LIMIT $3`,
		accountID, offset, r.itemsPerPage+1,
	)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch url: %w", err)
	}

	var nextCursor string

	if len(urls) > r.itemsPerPage {
		urls = urls[:r.itemsPerPage]
		nextCursor = strconv.Itoa(offset + r.itemsPerPage)
	}

	return urls, nextCursor, nil
}

func (r postgresV1) SetUrlState(ctx context.Context, accountID uint64, slug string, disabled bool) error {
	result, err := r.con.ExecContext(ctx, "UPDATE urls SET disabled=$3 WHERE slug=$1 AND account_id=$2", slug, accountID, disabled)
	if err != nil {
		return fmt.Errorf("failed to disable url: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}
