package testutil

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

const TestSecret = "test-secret"
const BadUUID = "bad-uuid"
const (
	UUID1 = "11111111-1111-4111-8111-111111111111"
	UUID2 = "22222222-2222-4222-8222-222222222222"
	UUID3 = "33333333-3333-4333-8333-333333333333"
	UUID4 = "44444444-4444-4444-8444-444444444444"
)

func MustUUID(t *testing.T, s string) pgtype.UUID {
	t.Helper()
	id, err := uuid.Parse(s)
	require.NoError(t, err)
	return pgtype.UUID{Bytes: id, Valid: true}
}
