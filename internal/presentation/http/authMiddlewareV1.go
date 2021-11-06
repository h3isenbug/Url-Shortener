package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/h3isenbug/url-shortener/internal/service/authentication"
	"github.com/h3isenbug/url-shortener/pkg/log"
)

type AuthMiddlewareV1 struct {
	basePresentationHandler

	authenticationService authentication.Service
}

func NewAuthMiddlewareV1(logger log.Logger, authenticationService authentication.Service) *AuthMiddlewareV1 {
	return &AuthMiddlewareV1{
		basePresentationHandler: basePresentationHandler{logger: logger},
		authenticationService:   authenticationService,
	}
}

func (m AuthMiddlewareV1) Intercept(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken := strings.TrimPrefix(r.Header.Get(headerAccessToken), "Bearer ")
		if accessToken == "" {
			m.sendResponseWithDefaultMessage(w, http.StatusUnauthorized)
			return
		}

		accountInfo, err := m.authenticationService.GetAccountInfoFromAccessToken(r.Context(), accessToken)
		if errors.Is(err, authentication.ErrExpiredToken) || errors.Is(err, authentication.ErrWrongCredentials) {
			m.sendResponseWithDefaultMessage(w, http.StatusUnauthorized)
			return
		}

		if err != nil {
			m.sendResponseWithDefaultMessage(w, http.StatusInternalServerError)
			return
		}

		ctxWithAccountInfo := context.WithValue(
			r.Context(),
			contextKeyAccountInfo,
			accountInfo,
		)

		next.ServeHTTP(w, r.WithContext(ctxWithAccountInfo))
	})
}
