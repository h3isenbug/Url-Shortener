package authentication

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/h3isenbug/url-shortener/internal/repository"
	"github.com/h3isenbug/url-shortener/internal/repository/account"
	refreshTokenRepository "github.com/h3isenbug/url-shortener/internal/repository/refreshToken"
	"github.com/h3isenbug/url-shortener/internal/types"
	"github.com/h3isenbug/url-shortener/pkg/log"
	"golang.org/x/crypto/bcrypt"
)

const (
	jwtSubject = "authentication"
)

const refreshTokenFamilyLength = 10

type jwtClaims struct {
	jwt.RegisteredClaims

	RefreshTokenID uint64 `json:"refreshTokenID,omitempty"`
	AccountID      uint64 `json:"accountID"`
}

type v1 struct {
	accountRepository      account.Repository
	refreshTokenRepository refreshTokenRepository.Repository
	logger                 log.Logger

	refreshTokenLength int

	refreshTokenLifespan time.Duration
	accessTokenLifespan  time.Duration

	accessTokenSecrets    map[string][]byte
	accessTokenCurrentKID string
}

func NewAuthenticationServiceV1(
	logger log.Logger,
	accountRepository account.Repository,
	refreshTokenRepository refreshTokenRepository.Repository,
	refreshTokenLength int,

	refreshTokenLifespan time.Duration,
	accessTokenLifespan time.Duration,

	accessTokenSecrets map[string][]byte,
	accessTokenCurrentKID string,
) Service {
	service := &v1{
		accountRepository:      accountRepository,
		refreshTokenRepository: refreshTokenRepository,
		logger:                 logger,
		refreshTokenLength:     refreshTokenLength,
		refreshTokenLifespan:   refreshTokenLifespan,
		accessTokenLifespan:    accessTokenLifespan,
		accessTokenSecrets:     accessTokenSecrets,
		accessTokenCurrentKID:  accessTokenCurrentKID,
	}
	if _, ok := accessTokenSecrets[accessTokenCurrentKID]; !ok {
		errorMessage := fmt.Sprintf(
			"auth token secrets are not configured correctly."+
				"no auth token secret was found with KID %s",
			accessTokenCurrentKID,
		)
		logger.Panic(errorMessage)
	}

	return service
}

func (s v1) getRandomEncodedBytes(count int) string {
	bytes := make([]byte, count)
	if _, err := rand.Read(bytes); err != nil {
		s.logger.Panic("failed to generate random bytes", map[string]interface{}{
			"errorMessage": err.Error(),
		})
	}

	return base64.RawURLEncoding.EncodeToString(bytes)
}

func (s v1) Login(ctx context.Context, email, password string) (*types.TokenPair, error) {
	acct, err := s.accountRepository.GetByEMail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrWrongCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account by email: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(acct.PasswordHash), []byte(password)); err != nil {
		return nil, ErrWrongCredentials
	}

	return s.generateTokenPair(ctx, acct.ID, nil)
}

func (s v1) generateTokenPair(ctx context.Context, accountID uint64, family *uint64) (*types.TokenPair, error) {
	refreshTokenText := s.getRandomEncodedBytes(s.refreshTokenLength)
	var refreshToken *types.RefreshToken
	var err error
	if family == nil {
		refreshToken, err = s.refreshTokenRepository.Create(
			ctx, accountID, refreshTokenText, s.refreshTokenLifespan)
	} else {
		refreshToken, err = s.refreshTokenRepository.CreateWithFamily(
			ctx, accountID, refreshTokenText, s.refreshTokenLifespan, *family)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	accessTokenText, err := s.generateAccessTokenForRefreshToken(accountID, refreshToken.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth token from refresh token(%d): %w", refreshToken.ID, err)
	}

	return &types.TokenPair{
		AccessToken:  accessTokenText,
		RefreshToken: refreshTokenText,
	}, nil

}

func (s v1) generateAccessTokenForRefreshToken(accountID, refreshTokenID uint64) (string, error) {
	now := jwt.NewNumericDate(time.Now().UTC())
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, &jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    types.ServiceName,
			Subject:   jwtSubject,
			Audience:  []string{types.ServiceName},
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenLifespan)),
			NotBefore: now,
			IssuedAt:  now,
			ID:        uuid.New().String(),
		},
		RefreshTokenID: refreshTokenID,
		AccountID:      accountID,
	})

	token.Header["kid"] = s.accessTokenCurrentKID

	signed, err := token.SignedString(s.accessTokenSecrets[s.accessTokenCurrentKID])
	if err != nil {
		s.logger.Panic("failed to generate signed jwt key", map[string]interface{}{"errorMessage": err.Error()})
		return "", fmt.Errorf("failed to generate signed jwt key: %w", err)
	}

	return signed, nil
}

func (s v1) parseAccessToken(tokenString string, skipClaimValidation bool) (*jwtClaims, error) {
	var claims jwtClaims

	parser := &jwt.Parser{
		ValidMethods:         []string{jwt.SigningMethodHS512.Alg()},
		SkipClaimsValidation: skipClaimValidation,
	}

	token, err := parser.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS512 {
			return nil, fmt.Errorf("expected signing method to be HS512, got %s", token.Method.Alg())
		}

		rawKID, ok := token.Header["kid"]
		if !ok {
			return nil, errors.New("no kid header was found in jwt")
		}

		kid, ok := rawKID.(string)
		if !ok {
			return nil, fmt.Errorf("expected KID header of jwt to be string, got %T", rawKID)
		}

		key, ok := s.accessTokenSecrets[kid]
		if !ok {
			return nil, fmt.Errorf("corresponding key for kid %s was not found", kid)
		}

		return key, nil
	})
	if token != nil && token.Valid {
		return &claims, nil
	}

	validationError, ok := err.(*jwt.ValidationError)
	if !ok {
		return nil, fmt.Errorf("failed to parse jwt token: %w", err)
	}

	if validationError.Errors&jwt.ValidationErrorExpired != 0 {
		return nil, ErrExpiredToken
	}

	if validationError.Errors&(jwt.ValidationErrorUnverifiable|jwt.ValidationErrorSignatureInvalid) != 0 {
		return nil, ErrTamperedToken
	}

	//all other claim related stuff
	if validationError.Errors&jwt.ValidationErrorClaimsInvalid > 0 {
		return nil, ErrWrongToken
	}

	return nil, fmt.Errorf("%w: unhandled jwt validation error: %d", ErrValidationFailed, validationError.Errors)
}
func (s v1) RenewTokens(ctx context.Context, oldAccessTokenString, refreshTokenString string) (*types.TokenPair, error) {
	claims, err := s.parseAccessToken(oldAccessTokenString, false)
	if err != nil {
		return nil, fmt.Errorf("failed to parse old auth token: %w", err)
	}

	refreshToken, err := s.refreshTokenRepository.Get(ctx, refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch refresh token: %w", err)
	}

	if refreshToken.ID != claims.RefreshTokenID {
		return nil, fmt.Errorf("%w: mismatching refresh token and auth token", ErrValidationFailed)
	}

	if refreshToken.ValidUntil.Before(time.Now().UTC()) {
		return nil, fmt.Errorf("%w: refresh token is expired", ErrExpiredToken)
	}

	if refreshToken.Compromised {
		return nil, fmt.Errorf("%w: compromised refresh token", ErrValidationFailed)
	}

	if refreshToken.Disabled {
		s.logger.Warn("a refresh token is getting used more than once. flagging refresh token as compromised", map[string]interface{}{
			"refreshTokenID": refreshToken.ID,
			"accountID":      refreshToken.AccountID,
		})
		if err := s.refreshTokenRepository.SetCompromisedState(ctx, refreshToken.Family); err != nil {
			s.logger.Error("failed to set refresh token family compromised flag", map[string]interface{}{
				"family":       refreshToken.Family,
				"errorMessage": err.Error(),
			})
		}
		return nil, fmt.Errorf("%w: attempted to reuse refresh token", ErrWrongCredentials)
	}

	tokenPair, err := s.generateTokenPair(ctx, refreshToken.AccountID, &refreshToken.Family)
	if err != nil {
		return nil, fmt.Errorf("failed to generate a new token pair: %w", err)
	}
	if err := s.refreshTokenRepository.Disable(ctx, refreshToken.ID); err != nil {
		return nil, fmt.Errorf("failed to disable old refresh token: %w", err)
	}

	return tokenPair, nil
}
func (s v1) Register(ctx context.Context, email, password string) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	err = s.accountRepository.Create(ctx, email, string(passwordHash))
	if errors.Is(err, repository.ErrUniquenessViolated) {
		return ErrEMailAlreadyUsed
	}
	if err != nil {
		return fmt.Errorf("failed to save new account: %w", err)
	}

	//TODO: stuff like sending verification email, sms, etc would be done here,
	// but authentication part of this project is just a MVP

	return nil
}

func (s v1) GetAccountInfoFromAccessToken(ctx context.Context, accessToken string) (*types.AccountInfo, error) {
	claims, err := s.parseAccessToken(accessToken, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get account info from auth token: %w", err)
	}

	// if access to other info about account is needed, it should be added here. do this IF it is really necessary.

	return &types.AccountInfo{ID: claims.AccountID}, nil
}
