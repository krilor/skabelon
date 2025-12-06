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
	// ErrInvalidFieldIdentifier is returned when a key identifier includes other characters than a-z and _.
	ErrInvalidFieldIdentifier = errors.New("invalid key identifier")
)

// RawJSON a struct for holding raw JSON messages while keeping order.
type RawJSON struct {
	fields []string
	values []json.RawMessage
}

// UnmarshalJSON implements json.Unmarshaler.
func (rj *RawJSON) UnmarshalJSON(data []byte) error {
	rawMap := make(map[string]json.RawMessage)

	err := json.Unmarshal(data, &rawMap)
	if err != nil {
		return fmt.Errorf("could not unmarshal raw json: %w", err)
	}

	fields := make([]string, 0, len(rawMap))
	values := make([]json.RawMessage, 0, len(rawMap))

	for field, value := range rawMap {
		if !regexp.MustCompile(`^[a-z_]*$`).MatchString(field) {
			return fmt.Errorf("%w: %s", ErrInvalidFieldIdentifier, field)
		}

		fields = append(fields, field)
		values = append(values, value)
	}

	rj.fields = fields
	rj.values = values

	return nil
}

// Fields returns the raw JSON field names.
func (rj *RawJSON) Fields() []string {
	return rj.fields
}

// Values returns the raw JSON values associated with the fields.
func (rj *RawJSON) Values() []json.RawMessage {
	return rj.values
}

func databaseType(jrm json.RawMessage) string {
	firstByte := jrm[0]
	switch firstByte {
	case '{': // object
		return "json"
	case '[': // array
		return "json"
	case 't': // true
		return "boolean"
	case 'f': // false
		return "boolean"
	case '"': // string
		return "text"
	case 'n': // null
		return "null"
	default:
		return "number"
	}
}

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

//nolint:funlen
func (s *Service) create(req *http.Request) (string, error) {
	ctx := req.Context()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", fmt.Errorf("no body: %w", err)
	}

	slog.InfoContext(ctx, "message", "body", string(body))

	rawJSON := new(RawJSON)

	err = json.Unmarshal(body, &rawJSON)
	if err != nil {
		return "", fmt.Errorf("could decode body: %w", err)
	}

	fields := rawJSON.Fields()
	values := rawJSON.Values()
	argNums := make([]string, len(fields))
	args := make([]any, len(fields))

	for idx, value := range values {
		switch databaseType(value) {
		case "null":
			// we cannot infer the actual type of a nulled field
			// by continuing we just set the arg to nil
			args[idx] = nil
		case "text":
			// text fields are quoted. We just remove the quotes and go on
			args[idx] = value[1 : len(value)-1]
		default:
			args[idx] = value
		}

		argNums[idx] = fmt.Sprintf("$%d", idx+1)
	}

	//nolint:gosec,unqueryvet // we need to do SQL string formatting for identifiers and splat since we want to generalize
	qry := fmt.Sprintf(`WITH _dbx_insert AS (
		INSERT INTO "skabelon"."resource" ( %[1]v )
		VALUES ( %[2]v )
		RETURNING "skabelon"."resource".*
	)
	SELECT
		coalesce(json_agg(_dbx_res)->0, 'null') AS _response
	FROM (
		SELECT * FROM _dbx_insert
	) _dbx_res;`,
		strings.Join(quoteIdentifiers(fields), ", "),
		strings.Join(argNums, ", "),
	)

	row := s.db.QueryRowContext(ctx, qry, args...)

	slog.InfoContext(ctx, "ready to query db", "query", qry)

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
