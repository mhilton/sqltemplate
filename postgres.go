package sqltemplate

import (
	"database/sql/driver"
	"fmt"
	"math"
	"strings"
	"time"
)

// PostgresLiteral formats the value v as a literal suitable for use in
// queries used with the PostgreSQL database.
//
// If v implements database/sql/driver.Valuer then Value() will be called
// before further processing.
//
// The literal form used for values of a specified type is:
//   nil
//     The SQL keyword NULL.
//   bool
//     Either the SQL keyword TRUE, or FALSE.
//   int, int64
//     The decimal value.
//   float64
//     If the value represents +Inf, -Inf or Nan then the literal will be
//     'Infinity', '-Infinity' or 'Nan' respectively. Otherwise the %g
//     encoding provided by fmt.Printf is used.
//   string
//     A string literal.
//   []byte
//     A bytea hex format literal, see
//     https://www.postgresql.org/docs/13/datatype-binary.html#id-1.5.7.12.9.
//   time.Time
//     A string literal containing the RFC3339 encoding of the time stamp.
//   Identifier
//     A quoted identifier, see
//     https://www.postgresql.org/docs/13/sql-syntax-lexical.html#SQL-SYNTAX-IDENTIFIERS.
func PostgresLiteral(v interface{}) (RawSQL, error) {
	if dv, ok := v.(driver.Valuer); ok {
		var err error
		v, err = dv.Value()
		if err != nil {
			return "", err
		}
	}
	switch v1 := v.(type) {
	case RawSQL:
		return v1, nil
	case Identifier:
		return RawSQL(`"` + strings.ReplaceAll(string(v1), `"`, `""`) + `"`), nil
	case bool:
		if v1 {
			return RawSQL("TRUE"), nil
		} else {
			return RawSQL("FALSE"), nil
		}
	case []byte:
		return RawSQL(fmt.Sprintf("'\\x%X'", v1)), nil
	case float64:
		if math.IsInf(v1, 1) {
			return RawSQL("'Infinity'"), nil
		}
		if math.IsInf(v1, -1) {
			return RawSQL("'-Infinity'"), nil
		}
		if math.IsNaN(v1) {
			return RawSQL("'NaN'"), nil
		}
		return RawSQL(fmt.Sprintf("%g", v1)), nil
	case int:
		return RawSQL(fmt.Sprintf("%d", v1)), nil
	case int64:
		return RawSQL(fmt.Sprintf("%d", v1)), nil
	case nil:
		return RawSQL("NULL"), nil
	case string:
		return RawSQL(`'` + strings.ReplaceAll(v1, `'`, `''`) + `'`), nil
	case time.Time:
		return RawSQL(`'` + v1.Format(time.RFC3339Nano) + `'`), nil
	}
	return "", fmt.Errorf("unknown type %T", v)
}
