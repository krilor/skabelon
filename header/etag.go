package header

import (
	"errors"
	"fmt"
	"regexp"
)

// ErrInvalidEtag is returned when an etag is invalid.
var ErrInvalidEtag = errors.New("invalid etag")

// etagPattern is used to validate etags
// This is a bit more restricted than what it probably should be.
// The spec allows "ASCII characters".
var etagPattern = regexp.MustCompile(`^[a-zA-Z0-9-_]*$`)

// ETag are the individual Etags in a Match header.
type ETag struct {
	weak bool
	etag string
}

// NewETag creates a new ETag.
func NewETag(weak bool, etag string) ETag {
	return ETag{weak: weak, etag: etag}
}

// String returns a string representation of the ETag.
func (e ETag) String() string {
	if e.weak {
		return "W/\"" + e.etag + "\""
	}

	return "\"" + e.etag + "\""
}

// IsWeak returns true if the ETag is weak.
func (e ETag) IsWeak() bool {
	return e.weak
}

// ETag returns the etag of the ETag.
func (e ETag) ETag() string {
	return e.etag
}

// Matching functions according to
// https://www.w3.org/Protocols/HTTP/1.1/rfc2616bis/issues/#i71

// MatchWeak returns true if the ETag matches the other ETag.
func (e ETag) MatchWeak(other ETag) bool {
	return e.etag == other.etag
}

// MatchStrong returns true if the ETag matches the other ETag.
func (e ETag) MatchStrong(other ETag) bool {
	if e.weak || other.weak {
		return false
	}

	return e.etag == other.etag
}

// ParseEtag creates a new ETag from a string.
func ParseEtag(etag string) (ETag, error) {
	if etag == "" {
		return ETag{}, fmt.Errorf("%w: empty etag value", ErrInvalidEtag)
	}

	if len(etag) < 3 { //nolint:mnd
		return ETag{}, fmt.Errorf("%w: etag is to short", ErrInvalidEtag)
	}

	weak := false
	if etag[0:2] == "W/" {
		weak = true
		etag = etag[2:]
	}

	if etag[0] != '"' && etag[len(etag)-1] != '"' {
		return ETag{}, fmt.Errorf("%w: missing quotes", ErrInvalidEtag)
	}

	etag = etag[1 : len(etag)-1]

	if !etagPattern.MatchString(etag) {
		return ETag{}, fmt.Errorf("%w: unwanted characters", ErrInvalidEtag)
	}

	return ETag{weak: weak, etag: etag}, nil
}
