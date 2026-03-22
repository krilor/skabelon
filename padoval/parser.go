package padoval

import (
	"encoding/json"
	"net/http"

	"github.com/krilor/skabelon/padoval/header"
)

// Parser is a middleware that parses the http.Request and passes it to the next [Handler].
// This is the main entry point for padoval.
func Parser(next Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, httpRequest *http.Request) {
		request, err := parse(httpRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		next.Serve(w, request)
	})
}

// parse returns a parsed request.
func parse(r *http.Request) (*Request, error) {
	method, err := MethodString(r.Method)
	if err != nil {
		return nil, err
	}

	headers, err := header.ParseRequest(r.Header)
	if err != nil {
		return nil, err
	}

	return &Request{
		Method:      method,
		Headers:     headers,
		Decoder:     json.NewDecoder(r.Body),
		httpRequest: r,
	}, nil
}
