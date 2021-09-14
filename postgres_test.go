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
	name:      "string with quotes",
	value:     "test 'string'",
	expectSQL: "'test ''string'''",
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
	name:      "false",
	value:     false,
	expectSQL: `FALSE`,
}, {
	name:      "bytes",
	value:     []byte("test"),
	expectSQL: `'\x74657374'`,
}, {
	name:      "float",
	value:     3.141592654,
	expectSQL: `3.141592654`,
}, {
	name:      "float Inf",
	value:     math.Inf(0),
	expectSQL: `'Infinity'`,
}, {
	name:      "float -Inf",
	value:     math.Inf(-1),
	expectSQL: `'-Infinity'`,
}, {
	name:      "float NaN",
	value:     math.NaN(),
	expectSQL: `'NaN'`,
}, {
	name:      "int",
	value:     0,
	expectSQL: `0`,
}, {
	name:      "int64",
	value:     int64(1e9),
	expectSQL: `1000000000`,
}, {
	name:      "time",
	value:     time.Date(2020, time.February, 2, 12, 30, 45, 300001000, time.UTC),
	expectSQL: `'2020-02-02T12:30:45.300001Z'`,
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
