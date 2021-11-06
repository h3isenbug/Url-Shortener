package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/h3isenbug/url-shortener/internal/service/authentication"
	"github.com/h3isenbug/url-shortener/pkg/log"
)

type authenticationV1 struct {
	basePresentationHandler

	authenticationService authentication.Service
}

func NewAuthenticationAPIV1(logger log.Logger, authenticationService authentication.Service) AuthenticationAPI {
	return &authenticationV1{
		basePresentationHandler: basePresentationHandler{
			logger: logger,
		},

		authenticationService: authenticationService,
	}
}

func (p authenticationV1) Login(w http.ResponseWriter, r *http.Request) {
	var request struct {
		EMail    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		p.sendResponseWithDefaultMessage(w, http.StatusBadRequest)
		return
	}

	tokens, err := p.authenticationService.Login(r.Context(), request.EMail, request.Password)
	if errors.Is(err, authentication.ErrWrongCredentials) {
		p.sendResponseWithDefaultMessage(w, http.StatusUnauthorized)
		return
	}
	if err != nil {
		p.sendResponseWithDefaultMessage(w, http.StatusInternalServerError)
		p.logger.Error("internal server error while handling login", map[string]interface{}{
			"username":     request.EMail,
			"errorMessage": err.Error(),
		})
		return
	}

	p.sendResponse(w, http.StatusOK, struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}{AccessToken: tokens.AccessToken, RefreshToken: tokens.RefreshToken})
}

func (p authenticationV1) RenewAccessToken(w http.ResponseWriter, r *http.Request) {
	var request struct {
		OldAccessToken string `json:"oldAccessToken"`
		RefreshToken   string `json:"refreshToken"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		p.sendResponseWithDefaultMessage(w, http.StatusBadRequest)
		return
	}

	newTokens, err := p.authenticationService.RenewTokens(r.Context(), request.OldAccessToken, request.RefreshToken)
	if errors.Is(err, authentication.ErrWrongCredentials) {
		p.sendResponseWithCustomMessage(w, http.StatusUnauthorized, "wrong credentials")
		return
	}
	if errors.Is(err, authentication.ErrExpiredToken) {
		p.sendResponseWithCustomMessage(w, http.StatusUnauthorized, "expired refresh token")
		return
	}
	if err != nil {
		p.sendResponseWithDefaultMessage(w, http.StatusInternalServerError)
		p.logger.Error("internal server error while handling auth token renewal", map[string]interface{}{
			"errorMessage": err.Error(),
		})
		return
	}

	p.sendResponse(w, http.StatusOK, struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}{AccessToken: newTokens.AccessToken, RefreshToken: newTokens.RefreshToken})
}

func (p authenticationV1) Register(w http.ResponseWriter, r *http.Request) {
	var request struct {
		EMail    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		p.sendResponseWithDefaultMessage(w, http.StatusBadRequest)
		return
	}

	err := p.authenticationService.Register(r.Context(), request.EMail, request.Password)
	if errors.Is(err, authentication.ErrEMailAlreadyUsed) {
		p.sendResponseWithCustomMessage(w, http.StatusBadRequest, "provided email address is already used")
		return
	}
	if err != nil {
		p.sendResponseWithDefaultMessage(w, http.StatusInternalServerError)
		p.logger.Error("internal server error while handling registration", map[string]interface{}{
			"email":        request.EMail,
			"errorMessage": err.Error(),
		})
		return
	}

	p.sendResponseWithDefaultMessage(w, http.StatusOK)
}
