package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/h3isenbug/url-shortener/internal/monitoring"
	"github.com/h3isenbug/url-shortener/internal/types"
	"github.com/h3isenbug/url-shortener/pkg/log"
)

const (
	headerAccessToken = "Authorization"
)

type ResponseWithMessage struct {
	Message string `json:"message"`
}

type basePresentationHandler struct {
	logger log.Logger
}

func (p basePresentationHandler) sendResponseWithCustomMessage(w http.ResponseWriter, statusCode int, message string) {
	p.sendResponse(w, statusCode, &ResponseWithMessage{message})
}

func (p basePresentationHandler) sendResponseWithDefaultMessage(w http.ResponseWriter, statusCode int) {
	p.sendResponse(w, statusCode, &ResponseWithMessage{http.StatusText(statusCode)})
}

func (p basePresentationHandler) sendResponse(w http.ResponseWriter, statusCode int, body interface{}) {
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		p.logger.Error("could not json encode or write response", map[string]interface{}{
			"statusCode":   statusCode,
			"errorMessage": err.Error(),
		})
	}
}

func GorillaMuxURLParamMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), contextKeyURLParams, mux.Vars(r))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getURLParams(r *http.Request) map[string]string {
	return r.Context().Value(contextKeyURLParams).(map[string]string)
}
func getAccountInfo(r *http.Request) *types.AccountInfo {
	return r.Context().Value(contextKeyAccountInfo).(*types.AccountInfo)
}

type teeResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newTeeResponseWriter(wrapped http.ResponseWriter) *teeResponseWriter {
	return &teeResponseWriter{
		ResponseWriter: wrapped,
		statusCode:     -1,
	}
}
func (w teeResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func GorillaHttpMetricsMiddleware(metricCollector monitoring.MetricCollector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tee := newTeeResponseWriter(w)
			startedAt := time.Now()
			next.ServeHTTP(tee, r)
			duration := time.Now().Sub(startedAt)

			path, _ := mux.CurrentRoute(r).GetPathTemplate()
			metricCollector.HttpResponseTime(
				r.Method,
				path,
				tee.statusCode,
				duration,
			)
		})
	}
}
