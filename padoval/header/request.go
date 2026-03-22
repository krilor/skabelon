package header

import (
	"fmt"
	"net/http"
)

// Request is a parsed HTTP request header.
type Request struct {
	// IfNoneMatch is the If-None-Match header
	IfNoneMatch *Match

	// IfMatch is the If-Match header
	IfMatch *Match
}

// ParseRequest takes a http.Header and returns a parsed Header.
// If the header is invalid, an error is returned.
// Headers that are not supported are silently ignored.
func ParseRequest(httpHeader http.Header) (*Request, error) {
	reqHeaders := Request{}

	if ifNoneMatch, ok := httpHeader[NameIfNoneMatch]; ok {
		if len(ifNoneMatch) > 1 {
			return &Request{}, fmt.Errorf("more than one %s header: %w", NameIfNoneMatch, ErrInvalidMatch)
		}

		if len(ifNoneMatch) == 1 {
			parsedIfNoneMatch, err := ParseMatch(ifNoneMatch[0])
			if err != nil {
				return &Request{}, err
			}

			reqHeaders.IfNoneMatch = &parsedIfNoneMatch
		}
	}

	if ifMatch, ok := httpHeader[NameIfMatch]; ok {
		if len(ifMatch) > 1 {
			return &Request{}, fmt.Errorf("more than one %s header: %w", NameIfMatch, ErrInvalidMatch)
		}

		if len(ifMatch) == 1 {
			parsedIfMatch, err := ParseMatch(ifMatch[0])
			if err != nil {
				return &Request{}, err
			}

			reqHeaders.IfMatch = &parsedIfMatch
		}
	}

	return &reqHeaders, nil
}
