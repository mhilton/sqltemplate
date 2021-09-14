package sqltemplate

import (
	"strings"
	"testing"
	"text/template/parse"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
)

func TestEscapeTreeIdempotent(t *testing.T) {
	text := `{{.}}
some text
{{if true}}{{.}}{{end}}
{{if true}}{{.}}{{else if true}}{{.}}{{else}}{{.}}{{end}}
{{range .}}{{.}}{{else}}{{.}}{{end}}
{{with "test"}}{{.}}{{end}}
{{with "test"}}{{.}}{{else}}{{end}}
`
	mt, err := parse.Parse("", text, "{{", "}}", nil)
	qt.Assert(t, err, qt.IsNil)

	t1 := mt[""]
	escapeTree(t1)
	t2 := t1.Copy()
	escapeTree(t2)

	qt.Check(t, t1, qt.CmpEquals(cmp.Comparer(parseTreeComparer)), t2)
}

func parseTreeComparer(t1, t2 *parse.Tree) bool {
	if t1 == t2 {
		return true
	}
	if t1.Name != t2.Name {
		return false
	}
	if t1.ParseName != t2.ParseName {
		return false
	}
	if t1.Mode != t2.Mode {
		return false
	}
	if len(t1.Root.Nodes) != len(t2.Root.Nodes) {
		return false
	}
	for i := range t1.Root.Nodes {
		if t1.Root.Nodes[i].String() != t2.Root.Nodes[i].String() {
			return false
		}
	}
	return true
}

func TestEscapeVarSettingPipe(t *testing.T) {
	tmpl, err := New("").Parse(`{{$v := printf "~%s~" . }}{{printf "<%s>" $v}}`)
	qt.Assert(t, err, qt.IsNil)

	var b strings.Builder
	err = tmpl.Execute(&b, "A")
	qt.Assert(t, err, qt.IsNil)
	qt.Check(t, b.String(), qt.Equals, `'<~A~>'`)
}
