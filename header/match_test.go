package header_test

import (
	"testing"

	"github.com/krilor/skabelon/header"
)

func TestParseMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		match   string
		wantErr bool
	}{
		{
			name:    "empty",
			match:   "",
			wantErr: true,
		},
		{
			name:    "missing quotes",
			match:   "abc",
			wantErr: true,
		},
		{
			name:    "match with unwanted characters",
			match:   "\"abc-123?\"",
			wantErr: true,
		},
		{
			name:    "match with quotes",
			match:   "\"abc\"",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := header.ParseMatch(tt.match)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
