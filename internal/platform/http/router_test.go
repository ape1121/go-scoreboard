package http

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteJSONSetsStatusCodeAndContentType(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()

	writeJSON(recorder, http.StatusAccepted, map[string]string{"status": "ok"})

	require.Equal(t, http.StatusAccepted, recorder.Code)
	require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	require.JSONEq(t, `{"status":"ok"}`, recorder.Body.String())
}

func TestHealthHandlerReturnsOKWhenDependencyIsReachable(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	handler := healthHandler(NewHealthService(stubPinger{}))
	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"status":"ok"}`, recorder.Body.String())
}

func TestHealthHandlerReturnsServiceUnavailableWhenDependencyFails(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	handler := healthHandler(NewHealthService(stubPinger{err: errors.New("db down")}))
	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	require.JSONEq(t, `{"status":"unhealthy"}`, recorder.Body.String())
}

func TestRequestIDMiddlewareSetsHeaderAndContext(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	handler := requestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"requestId": requestIDFromContext(r.Context())})
	}))
	handler.ServeHTTP(recorder, request)

	require.NotEmpty(t, recorder.Header().Get("X-Request-Id"))
	require.JSONEq(t, `{"requestId":"`+recorder.Header().Get("X-Request-Id")+`"}`, recorder.Body.String())
}

func TestRecovererMiddlewareReturnsStandardError(t *testing.T) {
	t.Parallel()

	var logs bytes.Buffer
	logger := log.New(&logs, "", 0)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/panic", nil).WithContext(
		context.WithValue(context.Background(), requestIDContextKey, "req-1"),
	)

	handler := recovererMiddleware(logger)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))
	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.JSONEq(t, `{"error":"internal server error"}`, recorder.Body.String())
	require.Contains(t, logs.String(), "req-1")
}

func TestRequestLoggerMiddlewareLogsRequest(t *testing.T) {
	t.Parallel()

	var logs bytes.Buffer
	logger := log.New(&logs, "", 0)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil).WithContext(
		context.WithValue(context.Background(), requestIDContextKey, "req-2"),
	)

	handler := requestLoggerMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r
		writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
	}))
	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusCreated, recorder.Code)
	require.Contains(t, logs.String(), "req-2")
	require.Contains(t, logs.String(), "status=201")
}

func TestRouterNotFoundUsesStandardErrorShape(t *testing.T) {
	t.Parallel()

	var logs bytes.Buffer
	router := NewRouter(Dependencies{
		Logger:        log.New(&logs, "", 0),
		HealthService: NewHealthService(stubPinger{}),
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/missing", nil)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNotFound, recorder.Code)
	require.JSONEq(t, `{"error":"route not found"}`, recorder.Body.String())
	require.NotEmpty(t, recorder.Header().Get("X-Request-Id"))
}

func TestRouterMethodNotAllowedUsesStandardErrorShape(t *testing.T) {
	t.Parallel()

	var logs bytes.Buffer
	router := NewRouter(Dependencies{
		Logger:        log.New(&logs, "", 0),
		HealthService: NewHealthService(stubPinger{}),
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodDelete, "/healthz", nil)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusMethodNotAllowed, recorder.Code)
	require.JSONEq(t, `{"error":"method not allowed"}`, recorder.Body.String())
}

func TestSkeletonSurroundingsRouteReturnsNotImplemented(t *testing.T) {
	t.Parallel()

	var logs bytes.Buffer
	router := NewRouter(Dependencies{
		Logger:        log.New(&logs, "", 0),
		HealthService: NewHealthService(stubPinger{}),
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/boards/board_1/scores/user_1/surroundings", nil)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNotImplemented, recorder.Code)
	require.JSONEq(t, `{"error":"endpoint not implemented"}`, recorder.Body.String())
}

func TestRequestIDHasExpectedFormat(t *testing.T) {
	t.Parallel()

	requestID := newRequestID()

	require.Len(t, requestID, 32)
	require.False(t, strings.Contains(requestID, "-"))
}

type stubPinger struct {
	err error
}

func (s stubPinger) Ping(context.Context) error {
	return s.err
}
