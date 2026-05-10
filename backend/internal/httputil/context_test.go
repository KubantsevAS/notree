package httputil_test

import (
	"context"
	"testing"

	"github.com/KubantsevAS/notree/backend/internal/httputil"
)

func TestGetUserIDFromCtx(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "user_id", "user-123")
		id, err := httputil.GetUserIDFromCtx(ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if id != "user-123" {
			t.Errorf("Expected user-123, got %s", id)
		}
	})

	t.Run("Missing in context", func(t *testing.T) {
		ctx := context.Background()
		_, err := httputil.GetUserIDFromCtx(ctx)
		if err == nil {
			t.Error("Expected error when user_id is missing")
		}
	})

	t.Run("Wrong type in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "user_id", 12345)
		_, err := httputil.GetUserIDFromCtx(ctx)
		if err == nil {
			t.Error("Expected error when user_id is of wrong type")
		}
	})
}
