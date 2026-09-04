package httputil

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSON(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{
			name:    "valid JSON",
			body:    `{"name":"Alpha"}`,
			wantErr: false,
		},
		{
			name:    "unknown field",
			body:    `{"unknown":"value"}`,
			wantErr: true,
		},
		{
			name:    "multiple JSON values",
			body:    `{"name":"Alpha"} {"name":"Beta"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(
				http.MethodPost,
				"/",
				strings.NewReader(tt.body),
			)
			response := httptest.NewRecorder()

			var actual payload
			err := DecodeJSON(response, request, &actual)

			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestDecodeJSONRejectsTooLargeBody(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	body := `{"name":"` + strings.Repeat("a", MaxRequestBodySize) + `"}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader(body),
	)
	response := httptest.NewRecorder()

	var actual payload
	err := DecodeJSON(response, request, &actual)

	if err == nil {
		t.Fatal("expected error for too large request body, got nil")
	}
}
