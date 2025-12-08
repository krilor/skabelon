package dbx

// Relation defines a relation (table or view) in the database.
type Relation struct {
	// Schema is the schema where the table or view is
	Schema string

	// Name is the name of the table or view
	Name string

	// Columns are the column names of the table or view
	Columns []string
}
