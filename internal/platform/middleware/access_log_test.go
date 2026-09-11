package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAccessLogPreservesResponse(t *testing.T) {
	handler := AccessLog(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"status":"created"}`))
	}))

	request := httptest.NewRequest(http.MethodPost, "/test", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, response.Code)
	}

	if response.Body.String() != `{"status":"created"}` {
		t.Fatalf("unexpected response body: %q", response.Body.String())
	}
}
