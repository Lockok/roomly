package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestRequireRoleAllowsMatchingRole(t *testing.T) {
	handler := RequireRole("admin")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	)

	ctx := context.WithValue(
		context.Background(),
		currentUserContextKey,
		CurrentUser{ID: uuid.New(), Role: "admin"},
	)

	request := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}
}

func TestRequireRoleRejectsNonMatchingRole(t *testing.T) {
	handler := RequireRole("admin")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("next handler must not be called")
		}),
	)

	ctx := context.WithValue(
		context.Background(),
		currentUserContextKey,
		CurrentUser{ID: uuid.New(), Role: "employee"},
	)

	request := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, response.Code)
	}
}
