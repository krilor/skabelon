//go:build dev

package dev

import (
	"net/http"
)

// HandleLiveReloadWebSocket registers the live reload websocket handler to mux when compiled for development
func HandleLiveReloadWebSocket(mux *http.ServeMux) {
	handleLiveReloadWebSocket(mux)
}

// LiveReloadScript is the live reload script when compiled for development
const LiveReloadScript = liveReloadScript
