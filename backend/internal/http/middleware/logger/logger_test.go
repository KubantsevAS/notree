package logger_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KubantsevAS/notree/backend/internal/http/middleware/logger"
	"github.com/stretchr/testify/require"
)

func TestLoggerMiddleware(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	h := logger.New(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("test body"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	res := httptest.NewRecorder()

	h.ServeHTTP(res, req)

	require.Equal(t, http.StatusAccepted, res.Code)
	require.Equal(t, "test body", res.Body.String())
}
