package room

import "testing"

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
