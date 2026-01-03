package dbx

import (
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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/krilor/skabelon/header"
	"github.com/lib/pq"
)

// CRUDHandler is a thingy for returning json from requests.
type CRUDHandler struct {
	http.Handler

	db  *pgxpool.Pool
	rel Relation
}

// TODO proper structure
// https://boldlygo.tech/posts/2024-01-08-error-handling/
// https://medium.com/@matryer/writing-middleware-in-golang-and-how-go-makes-it-so-much-fun-4375c1246e81
// https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/

// NewCRUDHandler returns a new Service.
//
//nolint:funlen
func NewCRUDHandler(db *pgxpool.Pool, relation Relation) *CRUDHandler {
	srv := CRUDHandler{ //nolint:exhaustruct
		db:  db,
		rel: relation,
	}

	mux := http.NewServeMux()
	// getOne endpoint based on id
	mux.HandleFunc("GET /{id}", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		id, err := strconv.Atoi(req.PathValue("id"))
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		str, etagStr, err := srv.getOne(req, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		etag := header.NewETag(false, etagStr)
		ifMatch := req.Header.Get("If-Match")

		if ifMatch != "" {
			match, err := header.ParseMatch(ifMatch)
			if err != nil {
				http.Error(w, "invalid If-Match header", http.StatusBadRequest)
				return
			}

			if match.MatchStrong(etag) {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("ETag", etag.String())
		w.Write([]byte(str)) //nolint:errcheck,gosec
	}))

	// updateOne endpoint based on id
	mux.HandleFunc("PATCH /{id}", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		id, err := strconv.Atoi(req.PathValue("id"))
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		str, err := srv.update(id, req)
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
	values []any
}

// NewRawJSONFromRequest returns a RawJSON a http request.
func NewRawJSONFromRequest(req *http.Request) (*RawJSON, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("no body: %w", err)
	}

	rawJSON := new(RawJSON)

	err = json.Unmarshal(body, &rawJSON)
	if err != nil {
		return nil, fmt.Errorf("could decode body: %w", err)
	}

	return rawJSON, nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (rj *RawJSON) UnmarshalJSON(data []byte) error {
	rawMap := make(map[string]json.RawMessage)

	err := json.Unmarshal(data, &rawMap)
	if err != nil {
		return fmt.Errorf("could not unmarshal raw json: %w", err)
	}

	fields := make([]string, 0, len(rawMap))
	values := make([]any, 0, len(rawMap))

	for field, value := range rawMap {
		if !regexp.MustCompile(`^[a-z_]*$`).MatchString(field) {
			return fmt.Errorf("%w: %s", ErrInvalidFieldIdentifier, field)
		}

		fields = append(fields, field)

		switch databaseType(value) {
		case "null":
			// we cannot infer the actual type of a nulled field
			// by continuing we just set the arg to nil
			values = append(values, nil)
		case "text":
			// text fields are quoted. We just remove the quotes and go on
			values = append(values, value[1:len(value)-1])
		default:
			values = append(values, value)
		}
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
func (rj *RawJSON) Values() []any {
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
func (s *CRUDHandler) getOne(req *http.Request, id int) (string, string, error) {
	ctx := req.Context()

	qry := fmt.Sprintf(`SELECT
			( select row_to_json(_obj) from (select %[1]s) as _obj ) as _response,
			_etag
		FROM "%[2]s"."%[3]s" _dbx  WHERE  "id" = $1 LIMIT 1`,
		strings.Join(prependIdentifier("_dbx", s.rel.Columns), " ,"),
		s.rel.Schema,
		s.rel.Name,
	)

	slog.InfoContext(ctx, "query prepped", "query", qry)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", "", fmt.Errorf("could not start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, qry, id)

	var (
		response string
		etag     string
	)

	switch err := row.Scan(&response, &etag); {
	case errors.Is(err, sql.ErrNoRows):
		return "", "", ErrNotFound
	case err == nil:
		return response, etag, nil
	default:
		panic(err)
	}
}

func (s *CRUDHandler) create(req *http.Request) (string, error) {
	ctx := req.Context()

	rawJSON, err := NewRawJSONFromRequest(req)
	if err != nil {
		return "", err
	}

	fields := rawJSON.Fields()
	argNums := make([]string, len(fields))

	for idx := range fields {
		argNums[idx] = fmt.Sprintf("$%d", idx+1)
	}

	qry := fmt.Sprintf(`WITH _dbx_insert AS (
		INSERT INTO "%[1]s"."%[2]s" ( %[4]s )
		VALUES ( %[5]s )
		RETURNING %[3]s
		)
		SELECT
		coalesce(json_agg(_dbx_res)->0, 'null') AS _response
		FROM (
			SELECT %[6]s FROM _dbx_insert
			) _dbx_res;`,
		s.rel.Schema,
		s.rel.Name,
		s.rel.returning(),
		strings.Join(quoteIdentifiers(fields), ", "),
		strings.Join(argNums, ", "),
		strings.Join(quoteIdentifiers(s.rel.Columns), ", "),
	)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("could not begin transaction: %w", err)
	}

	defer tx.Rollback(ctx)

	args := rawJSON.Values()
	row := tx.QueryRow(ctx, qry, args...)

	slog.InfoContext(ctx, "ready to query db", "query", qry)

	var response string

	err = row.Scan(&response)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}

		pgErr := &pq.Error{} //nolint:exhaustruct
		if errors.As(err, &pgErr) {
			slog.InfoContext(ctx, "pq error", "codename", pgErr.Code.Name(), "error", pgErr.Severity)
		}

		return "", fmt.Errorf("could not create resource: %w", err)
	}

	return response, nil
}

//nolint:funlen
func (s *CRUDHandler) update(id int, req *http.Request) (string, error) {
	ctx := req.Context()

	rawJSON, err := NewRawJSONFromRequest(req)
	if err != nil {
		return "", err
	}

	fields := rawJSON.Fields()
	setList := make([]string, len(fields))

	for idx, field := range fields {
		setList[idx] = fmt.Sprintf("\"%s\" = $%d", field, idx+1)
	}

	qry := fmt.Sprintf(`WITH _dbx_update AS (
		UPDATE "%[1]s"."%[2]s"
		SET %[4]v
		WHERE "id" = $%[5]d
		RETURNING %[3]s
		)
		SELECT
			coalesce(json_agg(_dbx_res)->0, 'null') AS _response
		FROM (
			SELECT %[6]s FROM _dbx_update
			) _dbx_res;`,
		s.rel.Schema,
		s.rel.Name,
		s.rel.returning(),
		strings.Join(setList, ", "),
		len(fields)+1,
		strings.Join(quoteIdentifiers(s.rel.Columns), ", "),
	)

	args := rawJSON.Values()
	args = append(args, id)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("could not begin transaction: %w", err)
	}

	row := tx.QueryRow(ctx, qry, args...)

	slog.InfoContext(ctx, "ready to query db", "query", qry)

	var response string

	err = row.Scan(&response)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}

		pgErr := &pq.Error{} //nolint:exhaustruct
		if errors.As(err, &pgErr) {
			slog.InfoContext(ctx, "pq error", "codename", pgErr.Code.Name(), "error", pgErr.Severity)
		}

		return "", fmt.Errorf("could not update resource: %w", err)
	}

	return response, nil
}

func quoteIdentifiers(s []string) []string {
	qs := make([]string, len(s))
	for i, v := range s {
		qs[i] = fmt.Sprintf(`"%s"`, v)
	}

	return qs
}

func prependIdentifier(identifier string, identifiers []string) []string {
	result := make([]string, len(identifiers))

	for i, v := range identifiers {
		result[i] = fmt.Sprintf(`"%s"."%s"`, identifier, v)
	}

	return result
}
