//go:build !dev

package dev

import (
	"net/http"
)

// HandleLiveReloadWebSocket registers noting when building for prod
func HandleLiveReloadWebSocket(mux *http.ServeMux) {
	return
}

// LiveReloadScript is empty when building for prod
const LiveReloadScript = ""
