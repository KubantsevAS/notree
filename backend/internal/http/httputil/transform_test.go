package httputil_test

import (
	"testing"

	"github.com/KubantsevAS/notree/backend/internal/http/httputil"
	"github.com/KubantsevAS/notree/backend/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestPgUUIDFromString(t *testing.T) {
	validStr := testutil.StringPtr(testutil.UUID1)
	invalidStr := testutil.StringPtr(testutil.BadUUID)

	tests := []struct {
		name    string
		input   *string
		wantErr bool
	}{
		{"Valid UUID", validStr, false},
		{"Invalid UUID", invalidStr, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := httputil.PgUUIDFromString(tt.input)
			require.Equal(t, tt.wantErr, err != nil)
		})
	}
}
