package httputil

import (
	"context"
	"errors"
)

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
