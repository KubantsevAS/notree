package router_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KubantsevAS/notree/backend/internal/config"
	"github.com/KubantsevAS/notree/backend/internal/http/middleware"
	"github.com/KubantsevAS/notree/backend/internal/router"
	"github.com/KubantsevAS/notree/backend/internal/testutil"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	authMark      = "auth-module"
	userMark      = "user-module"
	nodeMark      = "node-module"
	hierarchyMark = "hierarchy-module"
)

var registerMethods = []string{http.MethodGet, http.MethodPost}

type stubModule struct {
	path      string
	mark      string
	lastReqID string
}

func (m *stubModule) RegisterRoutes(r chi.Router) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, m.mark)
		if uid, ok := r.Context().Value(middleware.UserIDKey).(string); ok {
			_, _ = io.WriteString(w, ":"+uid)
		}
		m.lastReqID = chimw.GetReqID(r.Context())
	})
	for _, method := range registerMethods {
		r.Method(method, m.path, h)
	}
}

var protectedRoutes = []struct {
	path string
	mark string
}{
	{"/api/v1/stub-user", userMark},
	{"/api/v1/nodes/stub-node", nodeMark},
	{"/api/v1/nodes/stub-hierarchy", hierarchyMark},
}

const (
	testOrigin1 = "http://localhost:5173"
	testOrigin2 = "http://localhost:3000"
)

func testConfig() *config.Config {
	return &config.Config{
		Env: "test",
		JWT: config.JWTConfig{
			Secret: testutil.TestSecret,
		},
		CORSAllowedOriginsRaw: testOrigin1 + "," + testOrigin2,
	}
}

type testRouter struct {
	*chi.Mux
	authStub *stubModule
}

func newTestRouter(t *testing.T, log *slog.Logger) *testRouter {
	t.Helper()
	if log == nil {
		log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	authStub := &stubModule{path: "/stub-auth", mark: authMark}
	return &testRouter{
		Mux: router.New(
			testConfig(), log,
			authStub,
			&stubModule{path: "/stub-user", mark: userMark},
			&stubModule{path: "/stub-node", mark: nodeMark},
			&stubModule{path: "/stub-hierarchy", mark: hierarchyMark},
		),
		authStub: authStub,
	}
}

func TestRouteTable(t *testing.T) {
	r := newTestRouter(t, nil)

	var got []string
	err := chi.Walk(r, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		got = append(got, method+" "+route)
		return nil
	})
	require.NoError(t, err)

	want := []string{"GET /swagger/*"}
	paths := []string{"/api/v1/stub-auth"}
	for _, pr := range protectedRoutes {
		paths = append(paths, pr.path)
	}
	for _, p := range paths {
		for _, m := range registerMethods {
			want = append(want, m+" "+p)
		}
	}
	assert.ElementsMatch(t, want, got)
}

func TestAuthRoutesArePublic(t *testing.T) {
	r := newTestRouter(t, nil)

	rec := testutil.DoRequest(t, r, http.MethodGet, "/api/v1/stub-auth", nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, authMark, rec.Body.String())

	rec = testutil.DoRequest(t, r, http.MethodGet, "/api/v1/stub-auth", nil, testutil.AccessTokenCookie("garbage"))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestProtectedRoutesRequireAuth(t *testing.T) {
	testUserID := testutil.UUID1
	r := newTestRouter(t, nil)

	cases := []struct {
		name    string
		cookies []http.Cookie
	}{
		{name: "no cookie"},
		{name: "empty cookie value", cookies: []http.Cookie{testutil.AccessTokenCookie("")}},
		{name: "garbage token", cookies: []http.Cookie{testutil.AccessTokenCookie("not-a-jwt")}},
		{name: "wrong signing secret", cookies: []http.Cookie{testutil.AccessTokenCookie(testutil.MakeAccessToken(t, "wrong-secret", testUserID))}},
		{name: "expired token", cookies: []http.Cookie{testutil.AccessTokenCookie(testutil.MakeTokenWithClaims(t, testutil.TestSecret, jwt.MapClaims{
			"user_id": testUserID, "exp": time.Now().Add(-time.Minute).Unix(), "type": "access",
		}))}},
		{name: "refresh token instead of access", cookies: []http.Cookie{testutil.AccessTokenCookie(testutil.MakeTokenWithClaims(t, testutil.TestSecret, jwt.MapClaims{
			"user_id": testUserID, "exp": time.Now().Add(time.Hour).Unix(), "type": "refresh",
		}))}},
		{name: "token without user_id", cookies: []http.Cookie{testutil.AccessTokenCookie(testutil.MakeTokenWithClaims(t, testutil.TestSecret, jwt.MapClaims{
			"exp": time.Now().Add(time.Hour).Unix(), "type": "access",
		}))}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, pr := range protectedRoutes {
				rec := testutil.DoRequest(t, r, http.MethodGet, pr.path, nil, tc.cookies...)
				assert.Equal(t, http.StatusUnauthorized, rec.Code, "path: %s", pr.path)
			}
		})
	}
}

func TestProtectedRoutesWithValidToken(t *testing.T) {
	testUserID := testutil.UUID1
	r := newTestRouter(t, nil)
	token := testutil.MakeAccessToken(t, testutil.TestSecret, testUserID)

	for _, pr := range protectedRoutes {
		rec := testutil.DoRequest(t, r, http.MethodGet, pr.path, nil, testutil.AccessTokenCookie(token))
		assert.Equal(t, http.StatusOK, rec.Code, "path: %s", pr.path)
		assert.Equal(t, pr.mark+":"+testUserID, rec.Body.String(), "path: %s", pr.path)
	}
}

func TestCORS(t *testing.T) {
	r := newTestRouter(t, nil)

	t.Run("preflight from allowed origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/v1/stub-user", nil)
		req.Header.Set("Origin", testOrigin2)
		req.Header.Set("Access-Control-Request-Method", http.MethodPost)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, testOrigin2, rec.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
		assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), http.MethodPost)
	})

	t.Run("actual request from allowed origin", func(t *testing.T) {
		rec := testutil.DoRequest(t, r, http.MethodGet, "/api/v1/stub-auth",
			map[string]string{"Origin": testOrigin2})
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, testOrigin2, rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("foreign origin gets no CORS headers", func(t *testing.T) {
		rec := testutil.DoRequest(t, r, http.MethodGet, "/api/v1/stub-auth",
			map[string]string{"Origin": "https://evil.example"})
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestNotFound(t *testing.T) {
	testUserID := testutil.UUID1
	r := newTestRouter(t, nil)

	for _, path := range []string{
		"/", "/api/v1/unknown", "/api/v2/stub-user",
		"/api/v1/stub-auth/nope", "/api/v1/stub-user/extra/segment",
	} {
		rec := testutil.DoRequest(t, r, http.MethodGet, path, nil)
		assert.Equal(t, http.StatusNotFound, rec.Code, "path: %s", path)
	}

	rec := testutil.DoRequest(t, r, http.MethodGet, "/api/v1/nodes/nope", nil)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	token := testutil.MakeAccessToken(t, testutil.TestSecret, testUserID)
	rec = testutil.DoRequest(t, r, http.MethodGet, "/api/v1/nodes/nope", nil, testutil.AccessTokenCookie(token))
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestMethodNotAllowed(t *testing.T) {
	r := newTestRouter(t, nil)

	rec := testutil.DoRequest(t, r, http.MethodDelete, "/api/v1/stub-auth", nil)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestRequestID(t *testing.T) {
	r := newTestRouter(t, nil)

	t.Run("generated if absent", func(t *testing.T) {
		testutil.DoRequest(t, r, http.MethodGet, "/api/v1/stub-auth", nil)
		assert.NotEmpty(t, r.authStub.lastReqID)
	})

	t.Run("reused from incoming header", func(t *testing.T) {
		testutil.DoRequest(t, r, http.MethodGet, "/api/v1/stub-auth",
			map[string]string{chimw.RequestIDHeader: "fixed-id-42"})
		assert.Equal(t, "fixed-id-42", r.authStub.lastReqID)
	})
}

func TestURLFormat(t *testing.T) {
	r := newTestRouter(t, nil)

	rec := testutil.DoRequest(t, r, http.MethodGet, "/api/v1/stub-auth.json", nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, authMark, rec.Body.String())
}

func TestSwaggerServed(t *testing.T) {
	r := newTestRouter(t, nil)

	rec := testutil.DoRequest(t, r, http.MethodGet, "/swagger/index.html", nil)
	assert.Equal(t, http.StatusOK, rec.Code)
}
