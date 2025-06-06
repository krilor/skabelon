package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/krilor/skabelon/dev"
)

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
` + dev.LiveReloadScript + `
<script src="https://unpkg.com/htmx.org@2.0.4" integrity="sha384-HGfztofotfshcF7+8n44JQL2oJmowVChPTg48S+jvZoztPfvwD79OC/LTtG6dMp+" crossorigin="anonymous"></script>
</head>
<body>
    <h1>Hello from Go!</h1>
	<div id="parent-div">Its mess</div>
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
	mux := http.NewServeMux()
	mux.HandleFunc("/", serveHTML)

	// When compiling for development, register a websocket that can be used to live reload the frontend.
	dev.HandleLiveReloadWebSocket(mux)
	mux.Handle("/clicked", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "You clicked the button!")
	}))

	log.Println("Starting server on http://localhost:8080...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
