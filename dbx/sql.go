// Package dbx implements things for getting json directly from a sql.DB
package dbx

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
)

// Service is a thingy for returning json from requests.
type Service struct {
	http.Handler

	db *sql.DB
}

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

	srv.Handler = mux

	return &srv
}

// ErrNotFound is returned when a resource is not found.
var ErrNotFound = errors.New("not found")

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
