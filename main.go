// Package main is the base package for skabelon.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"text/template"
	"time"

	_ "github.com/lib/pq"

	"github.com/krilor/skabelon/dbx"
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

func dbConnection(ctx context.Context) (*sql.DB, error) {
	host := "localhost"
	port := 5432
	user := "postgres"
	password := "postgres_pwd" //nolint:gosec
	dbname := "postgres"
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("could not connect to db: %w", err)
	}

	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not pind db: %w", err)
	}

	return db, nil
}

// ServeHTTP implements http.Handler.
func (n *needNewName) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err := n.t.ExecuteTemplate(w, "index.tmpl", map[string]any{"Head": dev.LiveReloadHTML})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func start(ctx context.Context) error {
	n := newNeedNewName()
	mux := http.NewServeMux()
	mux.Handle("/", n)

	// When compiling for development, register a websocket that can be used to live reload the frontend.
	dev.HandleLiveReloadWebSocket(mux)

	mux.Handle("/clicked", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "You clicked the button!") //nolint:errcheck
	}))

	db, err := dbConnection(ctx)
	if err != nil {
		return err
	}
	defer db.Close() //nolint:errcheck

	service := dbx.NewService(db)
	mux.Handle("/resource/", LoggingMiddleware(http.StripPrefix("/resource", service)))

	slog.InfoContext(ctx, "Starting server on http://localhost:8080...")

	server := &http.Server{ //nolint:exhaustruct
		Handler:           mux,
		Addr:              ":8080",
		ReadHeaderTimeout: 3 * time.Second, //nolint:mnd
	}

	err = server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("ListenAndServe errored: %w", err)
	}

	return nil
}

// LoggingMiddleware logs the request using slog.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)
		slog.InfoContext(r.Context(), "message", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
	})
}

func main() {
	ctx := context.Background()

	err := start(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
