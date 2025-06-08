// Package main is the base package for skabelon.
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"text/template"
	"time"

	"github.com/krilor/skabelon/dev"
)

type needNewName struct {
	t *template.Template
}

// newNeedNewName returns a new NeedNewName.
func newNeedNewName() *needNewName {
	t := template.Must(template.ParseGlob("templates/*.tmpl"))
	return &needNewName{
		t: t,
	}
}

// ServeHTTP implements http.Handler.
func (n *needNewName) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := n.t.ExecuteTemplate(w, "index.tmpl", map[string]any{"Head": dev.LiveReloadHTML})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	ctx := context.Background()
	n := newNeedNewName()
	mux := http.NewServeMux()
	mux.Handle("/", n)

	// When compiling for development, register a websocket that can be used to live reload the frontend.
	dev.HandleLiveReloadWebSocket(mux)
	mux.Handle("/clicked", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "You clicked the button!") //nolint:errcheck
	}))

	slog.InfoContext(ctx, "Starting server on http://localhost:8080...")

	server := &http.Server{ //nolint:exhaustruct
		Handler:           mux,
		Addr:              ":8080",
		ReadHeaderTimeout: 3 * time.Second, //nolint:mnd
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
