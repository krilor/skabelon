package dbx

import (
	"strings"
)

// Relation defines a relation (table or view) in the database.
type Relation struct {
	// Schema is the schema where the table or view is
	Schema string

	// Name is the name of the table or view
	Name string

	// Columns are the column names of the table or view
	Columns []string
}

// String returns a string representation of the relation.
func (r Relation) String() string {
	return r.Schema + "." + r.Name + " (" + strings.Join(r.Columns, ", ") + ")"
}

// Returning constructs the string for the RETURNING clause
// of the sql query.
// The format is "schema"."name"."column", joined by ", ".
func (r Relation) returning() string {
	colRef := make([]string, len(r.Columns))
	for i, col := range r.Columns {
		colRef[i] = "\"" + r.Schema + "\"." + "\"" + r.Name + "\"." + "\"" + col + "\""
	}

	return strings.Join(colRef, ", ")
}
