package optional

import (
	"encoding/json"
	"testing"
)

func TestOptionalUnmarshalJSON(t *testing.T) {
	type payload struct {
		Floor Optional[int] `json:"floor"`
	}

	tests := []struct {
		name      string
		body      string
		wantSet   bool
		wantValue *int
	}{
		{
			name:    "field is absent",
			body:    `{}`,
			wantSet: false,
		},
		{
			name:      "field has value",
			body:      `{"floor": 3}`,
			wantSet:   true,
			wantValue: intPointer(3),
		},
		{
			name:    "field is null",
			body:    `{"floor": null}`,
			wantSet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual payload

			if err := json.Unmarshal([]byte(tt.body), &actual); err != nil {
				t.Fatalf("unmarshal JSON: %v", err)
			}

			if actual.Floor.Set != tt.wantSet {
				t.Fatalf("expected Set=%t, got %t", tt.wantSet, actual.Floor.Set)
			}

			if tt.wantValue == nil && actual.Floor.Value != nil {
				t.Fatalf("expected nil value, got %d", *actual.Floor.Value)
			}

			if tt.wantValue != nil {
				if actual.Floor.Value == nil {
					t.Fatal("expected non-nil value, got nil")
				}

				if *actual.Floor.Value != *tt.wantValue {
					t.Fatalf("expected value %d, got %d", *tt.wantValue, *actual.Floor.Value)
				}
			}
		})
	}
}

func intPointer(value int) *int {
	return &value
}
