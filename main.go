package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { // Allow all connections for development
		return true
	},
}

func serveDevLiveReloadWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade to websocket:", err)
		return
	}
	defer conn.Close()
	log.Println("DevLiveReload WebSocket connected")

	// Keep the connection alive, client reloads on close
	for {
		// You could read messages here if needed, but for simple live reload,
		// just keeping the connection open is often enough.
		// The client will attempt to reconnect if the server restarts (closing this conn).
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("DevLiveReload WebSocket disconnected:", err)
			break
		}
	}
}

const htmlContent = `
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Simple Go Web Page</title>
<style>
body { font-family: sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; background-color: #f0f0f0; }
h1 { color: #333; }
</style>
<script>
	conn = new WebSocket("ws://" + document.location.host + "/devlivereload");
	conn.onclose = function (evt) {
	console.log("Connection Closed")
	setTimeout(function () {
		location.reload();
	}, 2000);
	};
</script>
<script src="https://unpkg.com/htmx.org@2.0.4" integrity="sha384-HGfztofotfshcF7+8n44JQL2oJmowVChPTg48S+jvZoztPfvwD79OC/LTtG6dMp+" crossorigin="anonymous"></script>
</head>
<body>
    <h1>Hello from Go!</h1>
	<div id="parent-div">Its me</div>
	<button hx-post="/clicked"
    hx-trigger="click"
    hx-target="#parent-div"
    hx-swap="innerHTML">
    Click Me!
</button>
</body>
</html>
`

func serveHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, htmlContent)
}

func main() {
	http.HandleFunc("/", serveHTML)
	http.HandleFunc("/devlivereload", serveDevLiveReloadWS)
	http.Handle("/clicked", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "You clicked the button!")
	}))

	log.Println("Starting server on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
