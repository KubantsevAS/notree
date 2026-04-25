package httputil

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func PgTextFromString(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func PgUUIDFromString(s *string) (pgtype.UUID, error) {
	parsedUUID, err := uuid.Parse(*s)
	if err != nil {
		return pgtype.UUID{Valid: false}, err
	}

	return pgtype.UUID{Bytes: parsedUUID, Valid: true}, err
}

func RawMsgFromPtr(m *json.RawMessage) json.RawMessage {
	if m == nil {
		return nil
	}
	return *m
}
