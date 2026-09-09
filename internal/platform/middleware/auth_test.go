package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	platformauth "github.com/Lockok/roomly/internal/platform/auth"
	"github.com/google/uuid"
)

type fakeTokenParser struct {
	parse func(token string) (platformauth.Claims, error)
}

func (f fakeTokenParser) Parse(token string) (platformauth.Claims, error) {
	return f.parse(token)
}

func TestRequireAuth(t *testing.T) {
	userID := uuid.New()

	handler := RequireAuth(fakeTokenParser{
		parse: func(token string) (platformauth.Claims, error) {
			if token != "valid-token" {
				t.Fatalf("unexpected token: %q", token)
			}

			return platformauth.Claims{
				UserID: userID,
				Role:   "employee",
			}, nil
		},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentUser, ok := GetCurrentUser(r.Context())
		if !ok {
			t.Fatal("expected current user in context")
		}

		if currentUser.ID != userID {
			t.Fatalf("expected user ID %s, got %s", userID, currentUser.ID)
		}

		if currentUser.Role != "employee" {
			t.Fatalf("expected employee role, got %q", currentUser.Role)
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}
}

func TestRequireAuthRejectsMissingToken(t *testing.T) {
	handler := RequireAuth(fakeTokenParser{
		parse: func(string) (platformauth.Claims, error) {
			t.Fatal("token parser must not be called")
			return platformauth.Claims{}, nil
		},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler must not be called")
	}))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}
