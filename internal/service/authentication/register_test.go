package authentication_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/h3isenbug/url-shortener/internal/service/authentication"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistrationSuccessful(t *testing.T) {
	const email = "h.kalantari.1997@gmail.com"
	const password = "123456"

	authenticationService, accountRepo, _ := createSUT(t)

	accountRepo.EXPECT().Create(
		gomock.Any(), gomock.Eq(email), gomock.Any(),
	).Return(nil).Times(1)

	require.NoError(t, authenticationService.Register(context.Background(), email, password))
}

func TestRegistrationInUseEMail(t *testing.T) {
	const email = "h.kalantari.1997@gmail.com"
	const password = "123456"

	authenticationService, accountRepo, _ := createSUT(t)

	accountRepo.EXPECT().Create(
		gomock.Any(), gomock.Eq(email), gomock.Any(),
	).Return(authentication.ErrEMailAlreadyUsed).Times(1)

	if err := authenticationService.Register(context.Background(), email, password); assert.Error(t, err) {
		assert.ErrorIs(t, err, authentication.ErrEMailAlreadyUsed)
	}
}
