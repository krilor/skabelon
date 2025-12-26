package header_test

import (
	"testing"

	"github.com/krilor/skabelon/header"
)

func TestParseEtag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		etag    string
		wantErr bool
	}{
		{
			name:    "empty",
			etag:    "",
			wantErr: true,
		},
		{
			name:    "missing quotes",
			etag:    "abc",
			wantErr: true,
		},
		{
			name:    "weak etag with no quotes",
			etag:    "W/def",
			wantErr: true,
		},
		{
			name:    "etag with unwanted characters",
			etag:    "\"abc-123?\"",
			wantErr: true,
		},
		{
			name:    "etag with quotes",
			etag:    "\"abc\"",
			wantErr: false,
		},
		{
			name:    "valid etag",
			etag:    "\"abcdef123456\"",
			wantErr: false,
		},
		{
			name:    "valid weak etag",
			etag:    "W/\"abcdef123456\"",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := header.ParseEtag(tt.etag)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEtag() etag = %v,error = %v, wantErr %v", tt.etag, err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.String() != tt.etag {
					t.Errorf("ParseEtag() = %v, want %v", got.String(), tt.etag)
				}
			}
		})
	}
}

func TestEtagMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		first       string
		other       string
		matchWeak   bool
		matchStrong bool
	}{
		{
			name:        "same and strong",
			first:       "\"abcdef123456\"",
			other:       "\"abcdef123456\"",
			matchWeak:   true,
			matchStrong: true,
		},
		{
			name:        "same and both weak",
			first:       "W/\"abcdef123456\"",
			other:       "W/\"abcdef123456\"",
			matchWeak:   true,
			matchStrong: false,
		},
		{
			name:        "same and one weak",
			first:       "\"abcdef123456\"",
			other:       "W/\"abcdef123456\"",
			matchWeak:   true,
			matchStrong: false,
		},
		{
			name:        "different and weak",
			first:       "W/\"xyz123456\"",
			other:       "W/\"abcdef123456\"",
			matchWeak:   false,
			matchStrong: false,
		},
		{
			name:        "different and strong",
			first:       "\"xyz123456\"",
			other:       "\"abcdef123456\"",
			matchWeak:   false,
			matchStrong: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			first, _ := header.ParseEtag(tt.first)

			other, _ := header.ParseEtag(tt.other)
			if first.MatchWeak(other) != tt.matchWeak {
				t.Errorf("EtagMatch() weak = %v, want %v", first.MatchWeak(other), tt.matchWeak)
			}

			if first.MatchStrong(other) != tt.matchStrong {
				t.Errorf("EtagMatch() strong = %v, want %v", first.MatchStrong(other), tt.matchStrong)
			}
		})
	}
}
