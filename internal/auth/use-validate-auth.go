package auth

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
)

type ctxLoginKey string

var LoginKey ctxLoginKey = "login"

func UseValidateAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

		var updatedRequest *http.Request
		if err != nil || !token.Valid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), LoginKey, claims.Login)
		updatedRequest = r.WithContext(ctx)

		next.ServeHTTP(w, updatedRequest)
	})
}
