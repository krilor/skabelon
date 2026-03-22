package padoval

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/krilor/skabelon/padoval/header"
)

// Handler kinda looks like http.Handler, but it expects a parsed request.
type Handler interface {
	// Serve handles the request
	Serve(http.ResponseWriter, *Request)
}

// Method is a parsed HTTP method.
type Method int

//go:generate go run github.com/dmarkham/enumer -type=Method

const (
	MethodGET Method = iota
	MethodHEAD
	MethodPATCH
	MethodPOST
	MethodPUT
	MethodDELETE
	MethodQUERY
)

// Request is a parsed incoming HTTP request.
type Request struct {
	// Method is the HTTP method, such as GET or POST.
	Method Method

	// Header is the request header.
	Headers *header.Request

	// Decoder is the request body wrapped in [json.Decoder]
	Decoder *json.Decoder

	// httpRequest is the underlying http.Request
	httpRequest *http.Request
}

// URL returns URL from the underlying [http.Request].
func (r *Request) URL() *url.URL {
	return r.httpRequest.URL
}

// Host returns Host from the underlying [http.Request].
func (r *Request) Host() string {
	return r.httpRequest.Host
}

// Context returns the request's context.
func (r *Request) Context() context.Context {
	return r.httpRequest.Context()
}
