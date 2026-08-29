package jwt_test

import (
	"testing"

	"github.com/KubantsevAS/notree/backend/pkg/jwt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndParseAccessToken(t *testing.T) {
	userIdStr := "ebde9d75-dd29-4702-afde-1f93772f905d"
	parsedUUID, err := uuid.Parse(userIdStr)
	require.NoError(t, err)

	testUserID := pgtype.UUID{Bytes: parsedUUID, Valid: true}

	const testSecret = "secret-test-key"
	token, err := jwt.GenerateAccessToken(testUserID, testSecret)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	parsedUserID, err := jwt.ParseAccessToken(token, testSecret)
	require.NoError(t, err)
	require.Equal(t, userIdStr, parsedUserID)
}

func TestParseAccessToken_WrongSecret(t *testing.T) {
	userIdStr := "ebde9d75-dd29-4702-afde-1f93772f905d"
	parsedUUID, err := uuid.Parse(userIdStr)
	require.NoError(t, err)

	testUserID := pgtype.UUID{Bytes: parsedUUID, Valid: true}

	const correctSecret = "correct-test-key"
	token, err := jwt.GenerateAccessToken(testUserID, correctSecret)
	require.NoError(t, err)

	const wrongSecret = "wrong-test-key"
	_, err = jwt.ParseAccessToken(token, wrongSecret)
	require.Error(t, err)
}

func TestParseAccessToken_InvalidTokenString(t *testing.T) {
	const testSecret = "secret-test-key"
	_, err := jwt.ParseAccessToken("this.is.not.a.jwt", testSecret)
	require.Error(t, err)
}
