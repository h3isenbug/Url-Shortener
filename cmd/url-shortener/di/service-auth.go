package di

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/h3isenbug/url-shortener/internal/config"
	"github.com/h3isenbug/url-shortener/internal/repository/account"
	refreshTokenRepository "github.com/h3isenbug/url-shortener/internal/repository/refreshToken"
	"github.com/h3isenbug/url-shortener/internal/service/authentication"
	"github.com/h3isenbug/url-shortener/pkg/log"
	"gopkg.in/yaml.v2"
)

type accessTokenSecretsType map[string][]byte

func provideAccessTokenSecrets() (accessTokenSecretsType, error) {
	file, err := os.Open(config.Config.AccessTokenSecretFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read auth token secrets: %w", err)
	}
	defer file.Close()

	var rawSecretsFile struct {
		Secrets []struct {
			Name  string `yaml:"name"`
			Value string `yaml:"value"`
		} `yaml:"secrets"`
	}

	if err := yaml.NewDecoder(file).Decode(&rawSecretsFile); err != nil {
		return nil, fmt.Errorf("auth token secret file is not a valid yaml file: %w", err)
	}

	secrets := make(accessTokenSecretsType)

	for _, secret := range rawSecretsFile.Secrets {
		decoded, err := base64.StdEncoding.DecodeString(secret.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse secret %s: %w", secret.Name, err)
		}

		secrets[secret.Name] = decoded
	}

	return secrets, nil
}
func provideAuthenticationService(
	logger log.Logger,
	accountRepository account.Repository,
	refreshTokenRepository refreshTokenRepository.Repository,
	accessTokenSecrets accessTokenSecretsType,
) authentication.Service {
	return authentication.NewAuthenticationServiceV1(
		logger,
		accountRepository,
		refreshTokenRepository,
		config.Config.RefreshTokenLength,
		time.Duration(config.Config.RefreshTokenLifespanSeconds)*time.Second,
		time.Duration(config.Config.AccessTokenLifespanSeconds)*time.Second,
		accessTokenSecrets,
		config.Config.AccessTokenCurrentKID,
	)
}
