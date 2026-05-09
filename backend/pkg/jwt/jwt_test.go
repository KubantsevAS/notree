package jwt_test

import (
	"testing"

	"github.com/KubantsevAS/notree/backend/internal/httputil"
	"github.com/KubantsevAS/notree/backend/pkg/jwt"
)

func TestGenerateAndParseAccessToken(t *testing.T) {
	userIdStr := "ebde9d75-dd29-4702-afde-1f93772f905d"
	testUserID, err := httputil.PgUUIDFromString(&userIdStr)
	if err != nil {
		t.Fatal(err)
	}

	const testSecret = "secret-test-key"
	token, err := jwt.GenerateAccessToken(testUserID, testSecret)

	if err != nil {
		t.Fatalf("Token generation error: %v", err)
	}
	if token == "" {
		t.Fatal("Generated token should not be empty")
	}

	parsedUserID, err := jwt.ParseAccessToken(token, testSecret)
	if err != nil {
		t.Fatalf("Token parsing error: %v", err)
	}

	if parsedUserID != userIdStr {
		t.Errorf("Parsed user ID mismatch: expected %s, got %s", userIdStr, parsedUserID)
	}
}
