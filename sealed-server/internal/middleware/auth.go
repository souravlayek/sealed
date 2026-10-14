package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/souravlayek/sealed-server/pkg/auth"
	"github.com/souravlayek/sealed-server/utils"
)

type contextKey string

const UserClaimsKey contextKey = "userClaims"

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			log.Printf("authorization token is required")

			utils.SendErrorMessage(
				w,
				http.StatusUnauthorized,
				"authorization token is required",
			)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 || parts[0] != "Bearer" {

			utils.SendErrorMessage(
				w,
				http.StatusUnauthorized,
				"invalid authorization header",
			)
			return
		}

		claims, err := auth.ParseToken(parts[1])
		if err != nil {
			log.Printf("failed to parse authentication token: %v", err)

			utils.SendErrorMessage(
				w,
				http.StatusUnauthorized,
				"invalid or expired token",
			)
			return
		}

		ctx := context.WithValue(
			r.Context(),
			UserClaimsKey,
			claims,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
