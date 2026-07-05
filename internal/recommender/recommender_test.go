package recommender

import (
	"strings"
	"testing"
)

func TestReadLimited(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		max     int64
		want    string
		wantErr bool
	}{
		{
			name:  "under limit returns full bytes",
			input: "hello",
			max:   10,
			want:  "hello",
		},
		{
			name:  "exactly at limit returns full bytes",
			input: "hello",
			max:   5,
			want:  "hello",
		},
		{
			name:    "over limit returns error",
			input:   "hello world",
			max:     5,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readLimited(strings.NewReader(tt.input), tt.max)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Fatalf("got %q, want %q", string(got), tt.want)
			}
		})
	}
}
