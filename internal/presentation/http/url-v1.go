package http

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/h3isenbug/url-shortener/internal/repository"
	"github.com/h3isenbug/url-shortener/internal/service/url"
	"github.com/h3isenbug/url-shortener/internal/types"
	"github.com/h3isenbug/url-shortener/pkg/log"
)

type urlV1 struct {
	basePresentationHandler

	urlService url.Service
}

func NewUrlAPIV1(logger log.Logger, urlService url.Service) UrlAPI {
	return &urlV1{
		basePresentationHandler: basePresentationHandler{logger: logger},
		urlService:              urlService,
	}
}

func (p urlV1) GetOriginalUrl(w http.ResponseWriter, r *http.Request) {
	// ETag is used as a tracking mechanism here.

	slug := getURLParams(r)["slug"]
	newVisit := r.Header.Get("If-None-Match") == ""

	originalUrl, err := p.urlService.GetOriginalUrl(r.Context(), slug, newVisit)
	if errors.Is(err, repository.ErrNotFound) {
		p.sendResponseWithDefaultMessage(w, http.StatusNotFound)
		return
	}
	if err != nil {
		p.sendResponseWithDefaultMessage(w, http.StatusInternalServerError)
		return
	}

	w.Header().Set("ETag", base64.URLEncoding.EncodeToString([]byte(r.URL.String())))
	/*
	 *  If analytics is needed, use 302(client asks everytime), otherwise use 301
	 *    for better client-side performance.
	 */
	http.Redirect(
		w, r, originalUrl,
		http.StatusFound,
	)
}

func (p urlV1) CreateShortUrl(w http.ResponseWriter, r *http.Request) {
	var request struct {
		OriginalUrl string `json:"originalUrl"`
		Slug        string `json:"slug,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		p.sendResponseWithDefaultMessage(w, http.StatusBadRequest)
		return
	}

	slug, err := p.urlService.CreateShortUrl(r.Context(), request.OriginalUrl, request.Slug, getAccountInfo(r).ID)
	if errors.Is(err, repository.ErrUniquenessViolated) {
		p.sendResponseWithCustomMessage(w, http.StatusBadRequest, "requested slug is unavailable")
		return
	}
	if err != nil {
		p.sendResponseWithDefaultMessage(w, http.StatusInternalServerError)
		p.logger.Error("internal server error while saving new short url", map[string]interface{}{
			"errorMessage":        err.Error(),
			"originalUrl":         request.OriginalUrl,
			"recommendedShortUrl": request.Slug,
		})
		return
	}

	p.sendResponse(w, http.StatusCreated, struct {
		OriginalUrl string `json:"originalUrl"`
		Slug        string `json:"slug"`
	}{OriginalUrl: request.OriginalUrl, Slug: slug})
}

func (p urlV1) SetUrlState(w http.ResponseWriter, r *http.Request) {
	accountInfo := getAccountInfo(r)
	slug := getURLParams(r)["slug"]

	var request struct {
		Disabled bool `json:"disabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		p.sendResponseWithDefaultMessage(w, http.StatusBadRequest)
		return
	}

	err := p.urlService.SetUrlState(r.Context(), accountInfo.ID, slug, request.Disabled)
	if errors.Is(err, repository.ErrNotFound) {
		p.sendResponseWithDefaultMessage(w, http.StatusNotFound)
		return
	}
	if errors.Is(err, url.ErrNotAuthorized) {
		p.sendResponseWithDefaultMessage(w, http.StatusForbidden)
		return
	}
	if err != nil {
		p.sendResponseWithDefaultMessage(w, http.StatusInternalServerError)
		p.logger.Error("failed to disable short url", map[string]interface{}{
			"errorMessage": err.Error(),
			"accountID":    accountInfo.ID,
			"slug":         slug,
		})
		return
	}

	p.sendResponseWithDefaultMessage(w, http.StatusOK)
}

func (p urlV1) GetMyUrls(w http.ResponseWriter, r *http.Request) {
	accountInfo := getAccountInfo(r)
	cursor := r.URL.Query().Get("cursor")

	urls, nextCursor, err := p.urlService.GetAccountUrls(r.Context(), accountInfo.ID, cursor)
	if err != nil {
		p.sendResponseWithDefaultMessage(w, http.StatusInternalServerError)
		p.logger.Error("internal server error while getting user urls", map[string]interface{}{
			"errorMessage": err.Error(),
			"accountID":    accountInfo.ID,
		})
		return
	}

	p.sendResponse(w, http.StatusOK, &struct {
		Items      []types.Url `json:"items"`
		NextCursor string      `json:"nextCursor"`
	}{Items: urls, NextCursor: nextCursor})
}
