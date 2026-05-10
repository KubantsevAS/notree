package httputil_test

import (
	"testing"

	"github.com/KubantsevAS/notree/backend/internal/httputil"
)

func TestPgUUIDFromString(t *testing.T) {
	validStr := "ebde9d75-dd29-4702-afde-1f93772f905d"
	invalidStr := "not-a-uuid"

	tests := []struct {
		name    string
		input   *string
		wantErr bool
	}{
		{"Valid UUID", &validStr, false},
		{"Invalid UUID", &invalidStr, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := httputil.PgUUIDFromString(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("PgUUIDFromString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
