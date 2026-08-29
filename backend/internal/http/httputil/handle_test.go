package httputil_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/KubantsevAS/notree/backend/internal/http/httputil"
	"github.com/stretchr/testify/require"
)

type testPayload struct {
	Email string `validate:"required,email"`
	Age   int    `validate:"min=18"`
}

func TestHandleBody(t *testing.T) {
	t.Run("Valid body", func(t *testing.T) {
		jsonBody := `{"Email": "test@test.com", "Age": 25}`
		req := &http.Request{Body: io.NopCloser(strings.NewReader(jsonBody))}

		result, err := httputil.HandleBody[testPayload](req)
		require.NoError(t, err)
		require.Equal(t, "test@test.com", result.Email)
	})

	t.Run("Invalid JSON syntax", func(t *testing.T) {
		jsonBody := `{not a valid json`
		req := &http.Request{Body: io.NopCloser(strings.NewReader(jsonBody))}

		_, err := httputil.HandleBody[testPayload](req)
		require.Error(t, err)
	})

	t.Run("Validation failed", func(t *testing.T) {
		jsonBody := `{"Email": "not-an-email", "Age": 10}`
		req := &http.Request{Body: io.NopCloser(strings.NewReader(jsonBody))}

		_, err := httputil.HandleBody[testPayload](req)
		require.Error(t, err)
	})
}
