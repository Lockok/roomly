package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoveryReturnsInternalErrorAfterPanic(t *testing.T) {
	handler := Recovery(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		panic("unexpected error")
	}))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d",
			http.StatusInternalServerError,
			response.Code,
		)
	}

	expected := `{"code":"INTERNAL_ERROR","message":"internal server error"}` + "\n"
	if response.Body.String() != expected {
		t.Fatalf("unexpected response body: %q", response.Body.String())
	}
}