package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	platformauth "github.com/Lockok/roomly/internal/platform/auth"
	"github.com/Lockok/roomly/internal/platform/httputil"
	"github.com/google/uuid"
)

type authContextKey string

const currentUserContextKey authContextKey = "current_user"

type CurrentUser struct {
	ID   uuid.UUID
	Role string
}

type TokenParser interface {
	Parse(tokenString string) (platformauth.Claims, error)
}

type ActiveUserChecker func(ctx context.Context, userID uuid.UUID) (bool, error)

func RequireAuth(tokens TokenParser, activeUserCheck ...ActiveUserChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			parts := strings.Fields(header)

			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				httputil.WriteError(
					w,
					http.StatusUnauthorized,
					"UNAUTHORIZED",
					"authorization token is required",
				)
				return
			}

			claims, err := tokens.Parse(parts[1])
			if err != nil {
				code := "INVALID_TOKEN"
				message := "invalid access token"

				if errors.Is(err, platformauth.ErrExpiredToken) {
					code = "TOKEN_EXPIRED"
					message = "access token has expired"
				}

				httputil.WriteError(w, http.StatusUnauthorized, code, message)
				return
			}

			if len(activeUserCheck) > 0 {
				isActive, err := activeUserCheck[0](r.Context(), claims.UserID)
				if err != nil {
					httputil.WriteError(
						w,
						http.StatusServiceUnavailable,
						"AUTH_BACKEND_UNAVAILABLE",
						"unable to verify user status",
					)
					return
				}

				if !isActive {
					httputil.WriteError(
						w,
						http.StatusUnauthorized,
						"USER_INACTIVE",
						"user account is inactive",
					)
					return
				}
			}

			ctx := context.WithValue(
				r.Context(),
				currentUserContextKey,
				CurrentUser{
					ID:   claims.UserID,
					Role: claims.Role,
				},
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetCurrentUser(ctx context.Context) (CurrentUser, bool) {
	currentUser, ok := ctx.Value(currentUserContextKey).(CurrentUser)
	return currentUser, ok
}
