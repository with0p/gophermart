package auth

import (
	"context"
	"errors"
)

func GetLoginFromRequestContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(LoginKey).(string)
	if !ok {
		return "", errors.New("no login found")
	}

	return userID, nil
}
