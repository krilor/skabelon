package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/krilor/skabelon/dev"
)

type NeedNewName struct {
	t *template.Template
}

// NewNeedNewName returns a new NeedNewName
func NewNeedNewName() *NeedNewName {
	t := template.Must(template.ParseGlob("templates/*.tmpl"))
	return &NeedNewName{
		t: t,
	}
}

func (n *NeedNewName) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := n.t.ExecuteTemplate(w, "index.tmpl", map[string]any{"Head": dev.LiveReloadHTML})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	n := NewNeedNewName()
	mux := http.NewServeMux()
	mux.Handle("/", n)

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
