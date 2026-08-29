package testutil

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func UUIDFromString(s string) pgtype.UUID {
	var id pgtype.UUID
	err := id.Scan(s)
	if err != nil {
		panic(err)
	}
	return id
}

func UUIDFromStringT(t *testing.T, value string) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	if err := id.Scan(value); err != nil {
		t.Fatalf("pgtype.UUID scan failed: %v", err)
	}
	return id
}

func PgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func PgBool(b *bool) pgtype.Bool {
	if b == nil {
		return pgtype.Bool{Valid: false}
	}
	return pgtype.Bool{Bool: *b, Valid: true}
}

func PgTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}
