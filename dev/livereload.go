package dev

import (
	"fmt"
	"net/http"

	"golang.org/x/net/websocket"
)

// liveReloadWebSocketPath is the path where the server is exposing the live reload websocket
// when compiled for development.
const liveReloadWebSocketPath = "/devlivereload"

// liveReloadScript is the live reload script when compiled for development
const liveReloadScript = `<script>
	conn = new WebSocket("ws://" + document.location.host + "` + liveReloadWebSocketPath + `");
	conn.onclose = function (evt) {
	console.log("Connection Closed")
	setTimeout(function () {
		location.reload();
	}, 2000);
	};
</script>
`

func handleLiveReloadWebSocket(mux *http.ServeMux) {
	fmt.Println("Registering " + liveReloadWebSocketPath)
	mux.Handle(liveReloadWebSocketPath, websocket.Handler(devLiveReloadHandler))
	return
}

func devLiveReloadHandler(ws *websocket.Conn) {
	fmt.Println("Client connected to " + liveReloadWebSocketPath)
	// Handle incoming messages from the client...
	// Send messages to the client to initiate live reload...
	defer ws.Close()
	msg := make([]byte, 1024)
	for {
		_, err := ws.Read(msg)
		if err != nil {
			break
		}
	}
}
