package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
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

type stubPinger struct {
	err error
}

func (s stubPinger) Ping(context.Context) error {
	return s.err
}
