package sqltemplate

import (
	"database/sql"
	"math"
	"strings"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
)

var postgresLiteralTests = []struct {
	name      string
	value     interface{}
	expectSQL RawSQL
}{{
	name:      "nil",
	value:     nil,
	expectSQL: "NULL",
}, {
	name:      "simple string",
	value:     "test string",
	expectSQL: "'test string'",
}, {
	name:      "simple string pointer",
	value:     newString("test string"),
	expectSQL: "'test string'",
}, {
	name:      "string with quotes",
	value:     "test 'string'",
	expectSQL: "'test ''string'''",
}, {
	name:      "string with quotes pointer",
	value:     newString("test 'string'"),
	expectSQL: "'test ''string'''",
}, {
	name:      "nil string pointer",
	value:     (*string)(nil),
	expectSQL: "NULL",
}, {
	name:      "raw sql",
	value:     RawSQL("'; DROP TABLE users;"),
	expectSQL: "'; DROP TABLE users;",
}, {
	name:      "identifier",
	value:     Identifier("test identifier"),
	expectSQL: `"test identifier"`,
}, {
	name:      "identifier with quotes",
	value:     Identifier(`test "identifier"`),
	expectSQL: `"test ""identifier"""`,
}, {
	name:      "true",
	value:     true,
	expectSQL: `TRUE`,
}, {
	name:      "true pointer",
	value:     newBool(true),
	expectSQL: `TRUE`,
}, {
	name:      "false",
	value:     false,
	expectSQL: `FALSE`,
}, {
	name:      "false pointer",
	value:     newBool(false),
	expectSQL: `FALSE`,
}, {
	name:      "nil bool pointer",
	value:     (*bool)(nil),
	expectSQL: `NULL`,
}, {
	name:      "bytes",
	value:     []byte("test"),
	expectSQL: `'\x74657374'`,
}, {
	name:      "nil bytes",
	value:     []byte(nil),
	expectSQL: `NULL`,
}, {
	name:      "float",
	value:     3.141592654,
	expectSQL: `3.141592654`,
}, {
	name:      "float pointer",
	value:     newFloat(3.141592654),
	expectSQL: `3.141592654`,
}, {
	name:      "float Inf",
	value:     math.Inf(0),
	expectSQL: `'Infinity'`,
}, {
	name:      "float Inf pointer",
	value:     newFloat(math.Inf(0)),
	expectSQL: `'Infinity'`,
}, {
	name:      "float -Inf",
	value:     math.Inf(-1),
	expectSQL: `'-Infinity'`,
}, {
	name:      "float -Inf pointer",
	value:     newFloat(math.Inf(-1)),
	expectSQL: `'-Infinity'`,
}, {
	name:      "float NaN",
	value:     math.NaN(),
	expectSQL: `'NaN'`,
}, {
	name:      "float NaN pointer",
	value:     newFloat(math.NaN()),
	expectSQL: `'NaN'`,
}, {
	name:      "nil float pointer",
	value:     (*float64)(nil),
	expectSQL: `NULL`,
}, {
	name:      "int",
	value:     0,
	expectSQL: `0`,
}, {
	name:      "int pointer",
	value:     newInt(0),
	expectSQL: `0`,
}, {
	name:      "nil int pointer",
	value:     (*int)(nil),
	expectSQL: `NULL`,
}, {
	name:      "int64",
	value:     int64(1e9),
	expectSQL: `1000000000`,
}, {
	name:      "int64 pointer",
	value:     newInt64(1e9),
	expectSQL: `1000000000`,
}, {
	name:      "nil int64 pointer",
	value:     (*int64)(nil),
	expectSQL: `NULL`,
}, {
	name:      "time",
	value:     time.Date(2020, time.February, 2, 12, 30, 45, 300001000, time.UTC),
	expectSQL: `'2020-02-02T12:30:45.300001Z'`,
}, {
	name:      "time pointer",
	value:     newTime(time.Date(2020, time.February, 2, 12, 30, 45, 300001000, time.UTC)),
	expectSQL: `'2020-02-02T12:30:45.300001Z'`,
}, {
	name:      "niltime pointer",
	value:     (*time.Time)(nil),
	expectSQL: `NULL`,
}, {
	name: "valuer",
	value: sql.NullTime{
		Valid: true,
		Time:  time.Date(2020, time.February, 2, 12, 30, 45, 300005000, time.FixedZone("UTC-3", -3*60*60)),
	},
	expectSQL: `'2020-02-02T12:30:45.300005-03:00'`,
}}

func TestPostgresLiteral(t *testing.T) {
	for _, test := range postgresLiteralTests {
		t.Run(test.name, func(t *testing.T) {
			s, err := PostgresLiteral(test.value)
			qt.Assert(t, err, qt.IsNil)
			qt.Check(t, s, qt.Equals, test.expectSQL)
		})
	}
}

func TestPostgresLiteralInTemplate(t *testing.T) {
	tmpl, err := New("").Parse(`{{.}}`)
	qt.Assert(t, err, qt.IsNil)

	for _, test := range postgresLiteralTests {
		t.Run(test.name, func(t *testing.T) {
			sb := new(strings.Builder)
			err := tmpl.Execute(sb, test.value)
			qt.Assert(t, err, qt.IsNil)
			qt.Check(t, sb.String(), qt.Equals, string(test.expectSQL))
		})
	}
}

func TestPostgresLiteralUnknown(t *testing.T) {
	_, err := PostgresLiteral(make(chan bool))
	qt.Check(t, err, qt.ErrorMatches, `unknown type chan bool`)
}

func newBool(b bool) *bool {
	return &b
}

func newFloat(f float64) *float64 {
	return &f
}

func newInt(i int) *int {
	return &i
}

func newInt64(i int64) *int64 {
	return &i
}

func newString(s string) *string {
	return &s
}

func newTime(t time.Time) *time.Time {
	return &t
}
