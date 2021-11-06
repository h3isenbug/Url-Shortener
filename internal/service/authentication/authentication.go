package authentication

import (
	"context"
	"errors"
	"fmt"

	"github.com/h3isenbug/url-shortener/internal/types"
)

var (
	ErrValidationFailed = errors.New("validation error")
	ErrWrongCredentials = fmt.Errorf("%w: wrong credentials", ErrValidationFailed)
	ErrEMailAlreadyUsed = fmt.Errorf("%w: email already used", ErrValidationFailed)
	ErrExpiredToken     = fmt.Errorf("%w: token is expired", ErrValidationFailed)
	ErrTamperedToken    = fmt.Errorf("%w: token is tampered", ErrValidationFailed)
	ErrWrongToken       = fmt.Errorf("%w: token is malformed or was not meant for this purpose", ErrValidationFailed)
)

type Service interface {
	Login(ctx context.Context, email, password string) (*types.TokenPair, error)
	RenewTokens(ctx context.Context, oldAccessToken, refreshToken string) (*types.TokenPair, error)
	GetAccountInfoFromAccessToken(ctx context.Context, accessToken string) (*types.AccountInfo, error)
	Register(ctx context.Context, email, password string) error
}
