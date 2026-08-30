package httputil_test

import (
	"context"
	"testing"

	"github.com/KubantsevAS/notree/backend/internal/http/httputil"
	"github.com/KubantsevAS/notree/backend/internal/http/middleware"
	"github.com/stretchr/testify/require"
)

func TestGetUserIDFromCtx(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), middleware.UserIDKey, "user-123")
		id, err := httputil.GetUserIDFromCtx(ctx)
		require.NoError(t, err)
		require.Equal(t, "user-123", id)
	})

	t.Run("Missing in context", func(t *testing.T) {
		ctx := context.Background()
		_, err := httputil.GetUserIDFromCtx(ctx)
		require.Error(t, err)
	})

	t.Run("Wrong type in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), middleware.UserIDKey, 12345)
		_, err := httputil.GetUserIDFromCtx(ctx)
		require.Error(t, err)
	})
}
