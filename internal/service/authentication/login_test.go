package authentication_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/h3isenbug/url-shortener/internal/repository"
	"github.com/h3isenbug/url-shortener/internal/service/authentication"
	"github.com/h3isenbug/url-shortener/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginSuccessful(t *testing.T) {
	const email = "h.kalantari.1997@gmail.com"
	const password = "123456"

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	account := &types.Account{
		ID:           1,
		EMail:        email,
		PasswordHash: string(hash),
	}

	refreshToken := &types.RefreshToken{
		ID:          1,
		AccountID:   account.ID,
		ValidUntil:  time.Now().UTC().Add(time.Hour),
		Compromised: false,
		Disabled:    false,
		Family:      1,
		CreatedAt:   time.Now().UTC(),
	}
	authenticationService, accountRepo, refreshTokenRepo := createSUT(t)

	accountRepo.EXPECT().GetByEMail(
		gomock.Any(), gomock.Eq(email),
	).Return(account, nil).Times(1)

	refreshTokenRepo.EXPECT().Create(
		gomock.Any(), gomock.Eq(account.ID), gomock.Any(), time.Hour,
	).Return(refreshToken, nil).Times(1)

	tokenPair, err := authenticationService.Login(context.Background(), email, password)
	require.NoError(t, err)
	assert.Greater(t, len(tokenPair.AccessToken), 0, "Access token is empty")
	assert.Greater(t, len(tokenPair.RefreshToken), 0, "Refresh token is empty")
}

func TestLoginWrongEMail(t *testing.T) {
	const email = "h.kalantari.1997@gmail.com"
	const password = "123456"

	authenticationService, accountRepo, refreshTokenRepo := createSUT(t)

	accountRepo.EXPECT().GetByEMail(
		gomock.Any(), gomock.Eq(email),
	).Return(nil, repository.ErrNotFound).Times(1)

	refreshTokenRepo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	tokenPair, err := authenticationService.Login(context.Background(), email, password+"ASD")
	if assert.Error(t, err) {
		assert.ErrorIs(t, err, authentication.ErrWrongCredentials)
	}
	assert.Nil(t, tokenPair)
}

func TestLoginWrongPassword(t *testing.T) {
	const email = "h.kalantari.1997@gmail.com"
	const password = "123456"

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	account := &types.Account{
		ID:           1,
		EMail:        email,
		PasswordHash: string(hash),
	}

	authenticationService, accountRepo, refreshTokenRepo := createSUT(t)

	accountRepo.EXPECT().GetByEMail(
		gomock.Any(), gomock.Eq(email),
	).Return(account, nil).Times(1)

	refreshTokenRepo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	tokenPair, err := authenticationService.Login(context.Background(), email, password+"ASD")
	if assert.Error(t, err) {
		assert.ErrorIs(t, err, authentication.ErrWrongCredentials)
	}
	assert.Nil(t, tokenPair)
}
