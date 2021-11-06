package refreshToken

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/h3isenbug/url-shortener/internal/repository"
	"github.com/h3isenbug/url-shortener/internal/types"
	"github.com/jmoiron/sqlx"
)

type postgresV1 struct {
	con *sqlx.DB
}

func NewPostgresRepositoryV1(connection *sqlx.DB) Repository {
	return &postgresV1{con: connection}
}

func (r postgresV1) Create(ctx context.Context, accountID uint64, token string, lifespan time.Duration) (*types.RefreshToken, error) {
	var refreshToken types.RefreshToken
	err := r.con.GetContext(
		ctx, &refreshToken,
		`INSERT INTO refresh_tokens(account_id, token, valid_until) VALUES ($1, $2, $3)
					returning id, account_id, token, valid_until, compromised, disabled, family, created_at`,
		accountID, token, time.Now().UTC().Add(lifespan),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert refresh token: %w", err)
	}

	return &refreshToken, nil
}

func (r postgresV1) CreateWithFamily(ctx context.Context, accountID uint64, token string, lifespan time.Duration, family uint64) (*types.RefreshToken, error) {
	var refreshToken types.RefreshToken
	err := r.con.GetContext(
		ctx, &refreshToken,
		`INSERT INTO refresh_tokens(account_id, token, valid_until, family) VALUES ($1, $2, $3, $4)
 					returning id, account_id, token, valid_until, compromised, disabled, family, created_at`,
		accountID, token, time.Now().UTC().Add(lifespan), family,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert refresh token: %w", err)
	}

	return &refreshToken, nil
}

func (r postgresV1) Get(ctx context.Context, tokenString string) (*types.RefreshToken, error) {
	var token types.RefreshToken
	err := r.con.GetContext(ctx, &token, "SELECT id, account_id, token, valid_until, compromised, disabled, family FROM refresh_tokens WHERE token=$1", tokenString)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: refresh token not found(by token string)", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch refresh token: %w", err)
	}

	return &token, nil
}

func (r postgresV1) Disable(ctx context.Context, id uint64) error {
	result, err := r.con.ExecContext(ctx, "UPDATE refresh_tokens SET disabled=TRUE WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("failed to disable refresh token: %w", err)
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

func (r postgresV1) SetCompromisedState(ctx context.Context, family uint64) error {
	result, err := r.con.ExecContext(ctx, "UPDATE refresh_tokens SET compromised=TRUE WHERE family=$1", family)
	if err != nil {
		return fmt.Errorf("failed to mark refresh tokens as compromised: %w", err)
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
