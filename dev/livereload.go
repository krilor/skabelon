package dev

import (
	"html/template"
	"net/http"

	"golang.org/x/net/websocket"
)

// liveReloadWebSocketPath is the path where the server is exposing the live reload websocket
// when compiled for development.
const liveReloadWebSocketPath = "/devlivereload"

// liveReloadScript is the live reload script when compiled for development.
const liveReloadScript = `<script>
	conn = new WebSocket("ws://" + document.location.host + "` + liveReloadWebSocketPath + `");
	conn.onclose = function (evt) {
	console.log("Connection Closed")
	setTimeout(function () {
		location.reload();
	}, 1000);
	};
</script>
`

// LiveReloadHTML intended use it with html/template.
const LiveReloadHTML = template.HTML(liveReloadScript) //nolint:gosec

func handleLiveReloadWebSocket(mux *http.ServeMux) { //nolint:unused
	mux.Handle(liveReloadWebSocketPath, websocket.Handler(devLiveReloadHandler))
}

func devLiveReloadHandler(ws *websocket.Conn) { //nolint:unused
	// Handle incoming messages from the client...
	// Send messages to the client to initiate live reload...
	defer ws.Close() //nolint:errcheck
	const readSize = 1024
	msg := make([]byte, readSize)
	for {
		_, err := ws.Read(msg)
		if err != nil {
			break
		}
	}
}
