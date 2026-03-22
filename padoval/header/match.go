package header

import (
	"errors"
	"strings"
)

// ErrInvalidMatch is returned when a match header is invalid.
var ErrInvalidMatch = errors.New("invalid match header")

// Match is the header format used with If-None-Match and If-Match
// Match contains a list of values that can be matched.
// If the list is empty, it is interpreted as any/*
//
// Docs are here
//   - https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/If-None-Match
type Match struct {
	values []ETag
}

// NewMatch creates a new Match.
func NewMatch(value ...ETag) Match {
	return Match{
		values: value,
	}
}

// MatchWeak returns true if the ETag is matched using weak comparison.
func (m Match) MatchWeak(etag ETag) bool {
	if len(m.values) == 0 {
		return true
	}

	for _, v := range m.values {
		if v.MatchWeak(etag) {
			return true
		}
	}

	return false
}

// MatchStrong returns true if the ETag is matched using strong comparison.
func (m Match) MatchStrong(etag ETag) bool {
	if len(m.values) == 0 {
		return true
	}

	for _, v := range m.values {
		if v.MatchStrong(etag) {
			return true
		}
	}

	return false
}

// matchAny is the wildcard value.
const matchAny = "*"

// ErrEmptyHeader is returned when a header is empty.
var ErrEmptyHeader = errors.New("empty value")

// String returns a string representation of the Match.
func (m Match) String() string {
	if len(m.values) > 0 {
		etags := make([]string, len(m.values))
		for i, v := range m.values {
			etags[i] = v.String()
		}

		return strings.Join(etags, ", ")
	}

	return matchAny
}

// ParseMatch creates a new Match from a string.
func ParseMatch(value string) (Match, error) {
	if len(value) == 0 {
		return Match{}, errors.Join(ErrInvalidMatch, ErrEmptyHeader)
	}

	if value == matchAny {
		//nolint:exhaustruct
		return Match{}, nil
	}

	etags := strings.Split(value, ",")

	etagValues := make([]ETag, 0, len(etags))
	for _, etagStr := range etags {
		etagStr = strings.TrimSpace(etagStr)

		etag, err := ParseEtag(etagStr)
		if err != nil {
			return Match{}, errors.Join(ErrInvalidMatch, err)
		}

		etagValues = append(etagValues, etag)
	}

	return Match{values: etagValues}, nil
}
