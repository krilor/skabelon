// Package dbx implements things for getting json directly from a sql.DB
package dbx

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// Service is a thingy for returning json from requests.
type Service struct {
	http.Handler

	db *sql.DB
}

// TODO proper structure
// https://boldlygo.tech/posts/2024-01-08-error-handling/
// https://medium.com/@matryer/writing-middleware-in-golang-and-how-go-makes-it-so-much-fun-4375c1246e81
// https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/

// NewService returns a new Service.
func NewService(db *sql.DB) *Service {
	srv := Service{ //nolint:exhaustruct
		db: db,
	}

	mux := http.NewServeMux()
	// getOne endpoint based on id
	mux.HandleFunc("GET /{id}", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		id, err := strconv.Atoi(req.PathValue("id"))
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		str, err := srv.getOne(req.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(str)) //nolint:errcheck,gosec
	}))

	mux.HandleFunc("POST /", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		slog.InfoContext(req.Context(), "post")

		str, err := srv.create(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(str)) //nolint:errcheck,gosec
	}))

	srv.Handler = mux

	return &srv
}

var (
	// ErrNotFound is returned when a resource is not found.
	ErrNotFound = errors.New("not found")
	// ErrInvalidKeyIdentifier is returned when a key identifier includes other characters than a-z and _.
	ErrInvalidKeyIdentifier = errors.New("invalid key identifier")
)

// getOne returns a single resource.
func (s *Service) getOne(ctx context.Context, id int) (string, error) {
	//nolint:unqueryvet // We really do want to do select * here
	qry := `SELECT
		coalesce(json_agg(_dbx_res)->0, 'null') AS _response
	FROM ( SELECT * FROM "skabelon"."resource"  WHERE  "id" = $1 ) _dbx_res`

	row := s.db.QueryRowContext(ctx, qry, id)

	var response string

	switch err := row.Scan(&response); err {
	case sql.ErrNoRows:
		return "", ErrNotFound
	case nil:
		return response, nil
	default:
		panic(err)
	}
}

func (s *Service) create(req *http.Request) (string, error) {
	ctx := req.Context()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", fmt.Errorf("no body: %w", err)
	}

	slog.InfoContext(ctx, "message", "body", string(body))

	bodyMap := make(map[string]any)

	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		return "", fmt.Errorf("could decode body: %w", err)
	}

	keys := make([]string, 0, len(bodyMap))

	for key := range bodyMap {
		// check safe keys for sql
		if !regexp.MustCompile(`^[a-z_]*$`).MatchString(key) {
			return "", fmt.Errorf("%w: %s", ErrInvalidKeyIdentifier, key)
		}

		keys = append(keys, key)
	}

	slog.InfoContext(ctx, "message", "keys", strings.Join(keys, ","))

	// TODO the record definition just uses "text"

	//nolint:gosec,unqueryvet // we need to do SQL string formatting for identifiers and splat since we want to generalize
	qry := fmt.Sprintf(`WITH _dbx_insert AS (
		INSERT INTO "skabelon"."resource" ( %[1]v )
		SELECT dbx_body.%[2]v
		FROM ( SELECT $1::json AS _json_body ) dbx_request,
		LATERAL (
			SELECT %[1]v FROM json_to_record( dbx_request._json_body ) as _( %[3]v text)
		) dbx_body
		RETURNING "skabelon"."resource".*
	)
	SELECT
		coalesce(json_agg(_dbx_res)->0, 'null') AS _response
	FROM (
		SELECT * FROM _dbx_insert
	) _dbx_res`,
		strings.Join(quoteIdentifiers(keys), ","),
		strings.Join(quoteIdentifiers(keys), ", dbx_body."),
		strings.Join(quoteIdentifiers(keys), " text, "))

	slog.InfoContext(ctx, "ready to query db", "query", qry)

	row := s.db.QueryRowContext(ctx, qry, string(body))

	var response string

	switch err := row.Scan(&response); err {
	case sql.ErrNoRows:
		return "", ErrNotFound
	case nil:
		return response, nil
	default:
		return "", fmt.Errorf("could not create resource: %w", err)
	}
}

func quoteIdentifiers(s []string) []string {
	qs := make([]string, len(s))
	for i, v := range s {
		qs[i] = fmt.Sprintf(`"%s"`, v)
	}

	return qs
}
