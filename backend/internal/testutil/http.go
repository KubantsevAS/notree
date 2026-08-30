package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KubantsevAS/notree/backend/internal/http/dto"
	"github.com/stretchr/testify/require"
)

func NewJSONRequest(t *testing.T, method, url string, body any) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, url, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func AssertErrorJSON(t *testing.T, res *httptest.ResponseRecorder, expectedMsg string) {
	t.Helper()
	var payload dto.ErrorResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
	require.Equal(t, expectedMsg, payload.Error)
}

func AssertMessageJSON(t *testing.T, res *httptest.ResponseRecorder, expectedMsg string) {
	t.Helper()
	var payload dto.MessageResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
	require.Equal(t, expectedMsg, payload.Message)
}

func AssertAuthCookies(t *testing.T, res *httptest.ResponseRecorder, expectedNames ...string) []*http.Cookie {
	t.Helper()
	cookies := res.Result().Cookies()

	actual := make(map[string]struct{}, len(cookies))
	for _, cookie := range cookies {
		actual[cookie.Name] = struct{}{}
		require.True(t, cookie.HttpOnly, "cookie %s must be HttpOnly", cookie.Name)
	}

	for _, name := range expectedNames {
		_, ok := actual[name]
		require.True(t, ok, "missing cookie %s in %v", name, actual)
	}
	require.Len(t, cookies, len(expectedNames), "expected %d cookies, got %d", len(expectedNames), len(cookies))

	return cookies
}
