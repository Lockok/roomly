package room

import (
	"testing"

	"github.com/Lockok/roomly/internal/platform/optional"
)

func TestCreateInputValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   CreateInput
		wantErr bool
	}{
		{
			name: "valid request",
			input: CreateInput{
				Name:      "Aplha",
				Location:  "HQ",
				Capacity:  10,
				Equipment: []string{"tv", "whiteboard"},
			},
			wantErr: false,
		},

		{
			name: "empty name",
			input: CreateInput{
				Name:     "   ",
				Location: "HQ",
				Capacity: 10,
			},
			wantErr: true,
		},

		{
			name: "empty location",
			input: CreateInput{
				Name:     "Alpha",
				Location: " ",
				Capacity: 10,
			},
			wantErr: true,
		},

		{
			name: "zero capacity",
			input: CreateInput{
				Name:     "Alpha",
				Location: "HQ",
				Capacity: 0,
			},
			wantErr: true,
		},

		{
			name: "empty equipment item",
			input: CreateInput{
				Name:      "Alpha",
				Location:  "HQ",
				Capacity:  10,
				Equipment: []string{"tv", " "},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()

			if tt.wantErr && err == nil {
				t.Fatal("expected validation error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestUpdateInputValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   UpdateInput
		wantErr bool
	}{
		{
			name: "valid capacity update",
			input: UpdateInput{
				Capacity: optional.Optional[int]{
					Set:   true,
					Value: intPointer(12),
				},
			},
			wantErr: false,
		},
		{
			name:    "empty update",
			input:   UpdateInput{},
			wantErr: true,
		},
		{
			name: "null name",
			input: UpdateInput{
				Name: optional.Optional[string]{Set: true},
			},
			wantErr: true,
		},
		{
			name: "empty name",
			input: UpdateInput{
				Name: optional.Optional[string]{
					Set:   true,
					Value: stringPointer(" "),
				},
			},
			wantErr: true,
		},
		{
			name: "null floor is valid",
			input: UpdateInput{
				Floor: optional.Optional[int]{Set: true},
			},
			wantErr: false,
		},
		{
			name: "null equipment",
			input: UpdateInput{
				Equipment: optional.Optional[[]string]{Set: true},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()

			if tt.wantErr && err == nil {
				t.Fatal("expected validation error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func stringPointer(value string) *string {
	return &value
}

func intPointer(value int) *int {
	return &value
}
