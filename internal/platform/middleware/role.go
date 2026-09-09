package middleware

import (
	"net/http"

	"github.com/Lockok/roomly/internal/platform/httputil"
)

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			currentUser, ok := GetCurrentUser(r.Context())
			if !ok {
				httputil.WriteError(
					w,
					http.StatusUnauthorized,
					"UNAUTHORIZED",
					"authorization token is required",
				)
				return
			}

			for _, role := range roles {
				if currentUser.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			httputil.WriteError(
				w,
				http.StatusForbidden,
				"FORBIDDEN",
				"insufficient permissions",
			)
		})
	}
}
