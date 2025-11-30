// Package dbx implements things for getting json directly from a sql.DB
package dbx

import (
	"database/sql"
	"errors"
	"net/http"
)

// Service is a thingy for returning json from requests.
type Service struct {
	db *sql.DB
}

// NewService returns a new Service.
func NewService(db *sql.DB) *Service {
	return &Service{
		db: db,
	}
}

// ErrNotFound is returned when a resource is not found.
var ErrNotFound = errors.New("not found")

// ServeHTTP implements http.Handler.
func (s *Service) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	str, err := s.getOne(1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(str)) //nolint:errcheck,gosec
}

func (s *Service) getOne(id int) (string, error) {
	qry := `SELECT
		coalesce(json_agg(_dbx_res)->0, 'null') AS _response
	FROM ( SELECT * FROM "skabelon"."resource"  WHERE  "id" = $1 ) _dbx_res`

	row := s.db.QueryRow(qry, id)

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
