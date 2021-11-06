package authentication_test

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mockAccount "github.com/h3isenbug/url-shortener/internal/repository/account/mock"
	mockRefreshToken "github.com/h3isenbug/url-shortener/internal/repository/refreshToken/mock"
	"github.com/h3isenbug/url-shortener/internal/service/authentication"
	"github.com/h3isenbug/url-shortener/pkg/log"
	"github.com/stretchr/testify/require"
)

const refreshTokenLength = 30

func createSUT(t *testing.T) (authentication.Service, *mockAccount.MockRepository, *mockRefreshToken.MockRepository) {
	ctrl := gomock.NewController(t)

	accountRepo := mockAccount.NewMockRepository(ctrl)
	refreshTokenRepo := mockRefreshToken.NewMockRepository(ctrl)

	logger, err := log.NewZapLoggingService("")
	require.NoError(t, err)

	firstKey, err := base64.StdEncoding.DecodeString("9D9J0eqJPalytpmklvEK+2iDgCBI2m7bLUXtVLLJRq8AF4yE7QtJIg==")
	require.NoError(t, err)

	secondKey, err := base64.StdEncoding.DecodeString("tTvp/J03jQu8zkuJP6Vnrk/SuxTC9cWfOj2IsY+8XEzqwfyNm3alAA==")
	require.NoError(t, err)

	return authentication.NewAuthenticationServiceV1(
		logger,
		accountRepo,
		refreshTokenRepo,
		refreshTokenLength,
		time.Hour,
		time.Minute*10,
		map[string][]byte{
			"1": firstKey,
			"2": secondKey,
		},
		"2",
	), accountRepo, refreshTokenRepo
}
