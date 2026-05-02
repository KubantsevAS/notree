package httputil

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"
)

func GetUserPgUUIDFromCtx(ctx context.Context) (pgtype.UUID, error) {
	userIDContext, err := GetUserIDFromCtx(ctx)
	if err != nil {
		return pgtype.UUID{}, err
	}

	userID, err := PgUUIDFromString(&userIDContext)
	if err != nil {
		return pgtype.UUID{}, err
	}

	return userID, nil
}

func GetUserIDFromCtx(ctx context.Context) (string, error) {
	val := ctx.Value("user_id")
	if val == nil {
		return "", errors.New("User ID not found in context")
	}

	userID, ok := val.(string)
	if !ok {
		return "", errors.New("Invalid user id type in context")
	}

	return userID, nil
}
