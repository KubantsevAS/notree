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

func TestParseAccessToken_WrongSecret(t *testing.T) {
	userIdStr := "ebde9d75-dd29-4702-afde-1f93772f905d"
	testUserID, err := httputil.PgUUIDFromString(&userIdStr)
	if err != nil {
		t.Fatal(err)
	}

	const correctSecret = "correct-test-key"
	token, err := jwt.GenerateAccessToken(testUserID, correctSecret)
	if err != nil {
		t.Fatalf("Token generation error: %v", err)
	}

	const wrongSecret = "wrong-test-key"
	if _, err = jwt.ParseAccessToken(token, wrongSecret); err == nil {
		t.Fatalf("Expected error on parse with wrong secret")
	}
}

func TestParseAccessToken_InvalidTokenString(t *testing.T) {
	const testSecret = "secret-test-key"
	if _, err := jwt.ParseAccessToken("this.is.not.a.jwt", testSecret); err == nil {
		t.Error("Expected error for invalid token string")
	}
}
