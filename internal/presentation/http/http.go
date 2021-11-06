package http

import "net/http"

const (
	contextKeyAccountInfo = iota + 1
	contextKeyURLParams
)

type UrlAPI interface {
	CreateShortUrl(w http.ResponseWriter, r *http.Request)
	SetUrlState(w http.ResponseWriter, r *http.Request)

	GetMyUrls(w http.ResponseWriter, r *http.Request)
	GetOriginalUrl(w http.ResponseWriter, r *http.Request)
}

type AuthenticationAPI interface {
	Login(w http.ResponseWriter, r *http.Request)
	RenewAccessToken(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
}
