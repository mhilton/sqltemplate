// Package sqltemplate provides a template language to help generate
// correct SQL queries. It is built on the standard library's text/template
// (see https://golang.org/pkg/text/template) package, and uses the same
// template language.
//
// This package wraps the templates created by text/template such that the
// result of any pipeline is encoded using the sqlliteral function.
//
// Unlike the html/template package no attempt is made to derive semantic
// understanding of the template and encode values differently depending on
// where they are used. Templates in this package will always encode the
// same value in the same way regardless of context.
//
// # The sqlliteral function
//
// The sqlliteral template function must be a function of the form func(v
// interface{}) (RawSQL, error), the default implementation is
// PostgresLiteral.
//
// Implementations of sqlliteral must support any type that implements
// database/sql/driver.Valuer along with the types documented to make up
// the database/sql/driver.Value type. These are:
//
//	nil
//	int64
//	float64
//	bool
//	[]byte
//	string
//	time.Time
//
// Additional types may also be supported.
package sqltemplate
