package account

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/h3isenbug/url-shortener/internal/repository"
	"github.com/h3isenbug/url-shortener/internal/types"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type postgresV1 struct {
	con *sqlx.DB
}

func NewPostgresRepositoryV1(connection *sqlx.DB) Repository {
	return &postgresV1{con: connection}
}

func (r postgresV1) Get(ctx context.Context, id uint64) (*types.Account, error) {
	var account types.Account
	err := r.con.GetContext(ctx, &account, "SELECT id, email, password_hash FROM accounts WHERE id=$1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("an account with id=%d was not found: %w", id, err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch account by id(%d): %w", id, err)
	}

	return &account, nil
}

func (r postgresV1) GetByEMail(ctx context.Context, email string) (*types.Account, error) {
	var account types.Account
	err := r.con.GetContext(ctx, &account, "SELECT id, email, password_hash FROM accounts WHERE email=$1", email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("an account with email=%s was not found: %w", email, err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch account by email(%s): %w", email, err)
	}

	return &account, nil
}

func (r postgresV1) Create(ctx context.Context, email, password string) error {
	_, err := r.con.ExecContext(ctx, "INSERT INTO accounts (email, password_hash) VALUES($1, $2)", email, password)
	if err == nil {
		return nil
	}

	if pqError, ok := err.(*pq.Error); ok && pqError.Code.Name() == "unique_violation" {
		return fmt.Errorf("%w: an account with the given email already exists", repository.ErrUniquenessViolated)
	}

	return fmt.Errorf("failed to insert account(%s): %w", email, err)
}
