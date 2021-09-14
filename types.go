package sqltemplate

// An Identifier holds a value that should be formatted as an identifier in
// the SQL output.
type Identifier string

// A RawSQL value contains part of an SQL query that will be inserted into
// the template output verbatim.
type RawSQL string
