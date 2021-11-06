package types

import "time"

const ServiceName = "url-shortener"

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type Account struct {
	ID           uint64 `db:"id"`
	EMail        string `db:"email"`
	PasswordHash string `db:"password_hash"`
}

type AccountInfo struct {
	ID uint64
}

type RefreshToken struct {
	ID          uint64    `db:"id"`
	AccountID   uint64    `db:"account_id"`
	Token       string    `db:"token"`
	ValidUntil  time.Time `db:"valid_until"`
	Compromised bool      `db:"compromised"`
	Disabled    bool      `db:"disabled"`
	Family      uint64    `db:"family"`
	CreatedAt   time.Time `db:"created_at"`
}
