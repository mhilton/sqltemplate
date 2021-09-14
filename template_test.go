package sqltemplate

import (
	"embed"
	"sort"
	"strings"
	"testing"
	"text/template"

	qt "github.com/frankban/quicktest"
)

//go:embed testdata/*
var testData embed.FS

func TestNew(t *testing.T) {
	tmpl, err := New("test-1").Parse(`{{.}}`)
	qt.Assert(t, err, qt.IsNil)
	qt.Check(t, tmpl.Name(), qt.Equals, "test-1")

	var sb strings.Builder
	err = tmpl.Execute(&sb, "test-2")
	qt.Assert(t, err, qt.IsNil)
	qt.Check(t, sb.String(), qt.Equals, "'test-2'")
}

func TestMust(t *testing.T) {
	tmpl := Must(New("test").Parse(""))
	qt.Check(t, tmpl.Name(), qt.Equals, "test")
	qt.Check(t, func() { Must(New("").Parse("{{")) }, qt.PanicMatches, `template: :1: unclosed action`)
}

func TestParseFS(t *testing.T) {
	tmpl, err := ParseFS(testData, "testdata/*.tmpl")
	qt.Assert(t, err, qt.IsNil)
	qt.Check(t, tmpl.Name(), qt.Equals, "a.tmpl")
	var names []string
	for _, tmpl := range tmpl.Templates() {
		names = append(names, tmpl.Name())
	}
	sort.Strings(names)
	qt.Check(t, names, qt.DeepEquals, []string{
		"B",
		"a.tmpl",
		"b.tmpl",
		"c.tmpl",
	})

	_, err = ParseFS(testData)
	qt.Check(t, err, qt.ErrorMatches, "sqltemplate: no patterns provided in call to ParseFS")

	_, err = ParseFS(testData, "")
	qt.Check(t, err, qt.ErrorMatches, "sqltemplate: pattern matches no files: ``")

	_, err = ParseFS(testData, "\\")
	qt.Check(t, err, qt.ErrorMatches, "syntax error in pattern")
}

func TestParseFiles(t *testing.T) {
	tmpl, err := ParseFiles("testdata/a.tmpl", "testdata/b.tmpl", "testdata/c.tmpl")
	qt.Assert(t, err, qt.IsNil)
	qt.Check(t, tmpl.Name(), qt.Equals, "a.tmpl")
	var names []string
	for _, tmpl := range tmpl.Templates() {
		names = append(names, tmpl.Name())
	}
	sort.Strings(names)
	qt.Check(t, names, qt.DeepEquals, []string{
		"B",
		"a.tmpl",
		"b.tmpl",
		"c.tmpl",
	})

	_, err = ParseFiles()
	qt.Check(t, err, qt.ErrorMatches, "sqltemplate: no files named in call to ParseFiles")
}

func TestParseGlob(t *testing.T) {
	tmpl, err := ParseGlob("testdata/*.tmpl")
	qt.Assert(t, err, qt.IsNil)
	qt.Check(t, tmpl.Name(), qt.Equals, "a.tmpl")
	var names []string
	for _, tmpl := range tmpl.Templates() {
		names = append(names, tmpl.Name())
	}
	sort.Strings(names)
	qt.Check(t, names, qt.DeepEquals, []string{
		"B",
		"a.tmpl",
		"b.tmpl",
		"c.tmpl",
	})

	_, err = ParseGlob("")
	qt.Check(t, err, qt.ErrorMatches, "sqltemplate: pattern matches no files: ``")

	_, err = ParseGlob("\\")
	qt.Check(t, err, qt.ErrorMatches, "syntax error in pattern")
}

func TestTemplateAddParseTree(t *testing.T) {
	var st Template
	tt, err := template.New("text").Parse("{{.}}")
	qt.Assert(t, err, qt.IsNil)
	_, err = st.AddParseTree("", tt.Tree)
	qt.Assert(t, err, qt.IsNil)

	var sb1, sb2 strings.Builder
	err = tt.Execute(&sb1, "test")
	qt.Assert(t, err, qt.IsNil)
	err = st.Execute(&sb2, "test")
	qt.Assert(t, err, qt.IsNil)

	qt.Check(t, sb1.String(), qt.Equals, "test")
	qt.Check(t, sb2.String(), qt.Equals, "'test'")
}

func TestTemplateClone(t *testing.T) {
	var t1 Template
	_, err := t1.Parse(`A{{.}}A`)
	qt.Assert(t, err, qt.IsNil)

	t2, err := Must(t1.Clone()).Parse(`B{{.}}B`)
	qt.Assert(t, err, qt.IsNil)

	var sb1, sb2 strings.Builder
	err = t1.Execute(&sb1, "test-1")
	qt.Assert(t, err, qt.IsNil)
	qt.Check(t, sb1.String(), qt.Equals, "A'test-1'A")
	err = t2.Execute(&sb2, "test-1")
	qt.Assert(t, err, qt.IsNil)
	qt.Check(t, sb2.String(), qt.Equals, "B'test-1'B")
}

func TestTemplateDefinedTemplates(t *testing.T) {
	var tmpl Template
	qt.Check(t, tmpl.DefinedTemplates(), qt.Equals, "")

	_, err := tmpl.ParseFiles("testdata/a.tmpl", "testdata/b.tmpl", "testdata/c.tmpl")
	qt.Assert(t, err, qt.IsNil)
	qt.Check(t, tmpl.DefinedTemplates(), qt.Matches, `; defined templates are: ".*", ".*", ".*", ".*"`)
}

func TestTemplateDelims(t *testing.T) {
	tmpl, err := new(Template).Delims("<<", ">>").Parse(`<<.>>`)
	qt.Assert(t, err, qt.IsNil)

	var b strings.Builder
	err = tmpl.Execute(&b, "test")
	qt.Assert(t, err, qt.IsNil)
	qt.Check(t, b.String(), qt.Equals, "'test'")

}

func TestTemplateExecute(t *testing.T) {
	var b strings.Builder
	err := new(Template).Execute(&b, nil)
	qt.Check(t, err, qt.ErrorMatches, `sqltemplate: "" is an incomplete or empty template`)

	tmpl, err := new(Template).Parse(`{{.}}`)
	qt.Assert(t, err, qt.IsNil)
	err = tmpl.Execute(&b, "test")
	qt.Assert(t, err, qt.IsNil)
	qt.Check(t, b.String(), qt.Equals, "'test'")
}

func TestTemplateExecuteTemplate(t *testing.T) {
	var b strings.Builder
	err := new(Template).ExecuteTemplate(&b, "test-template", nil)
	qt.Check(t, err, qt.ErrorMatches, `sqltemplate: no template "test-template" associated with template ""`)

	tmpl, err := new(Template).Parse(`{{define "test-template"}}-{{.}}-{{end}}`)
	qt.Assert(t, err, qt.IsNil)
	err = tmpl.ExecuteTemplate(&b, "test-template", "test")
	qt.Assert(t, err, qt.IsNil)
	qt.Check(t, b.String(), qt.Equals, "-'test'-")
}

func TestTemplateFuncs(t *testing.T) {
	tmpl := new(Template).Funcs(FuncMap{
		"testf": func() string { return "test value" },
	})
	tmpl, err := tmpl.Parse(`{{ testf }}`)
	qt.Assert(t, err, qt.IsNil)

	var b strings.Builder
	err = tmpl.Execute(&b, nil)
	qt.Assert(t, err, qt.IsNil)
	qt.Check(t, b.String(), qt.Equals, "'test value'")
}

func TestTemplateLookup(t *testing.T) {
	tmpl := new(Template).Lookup("test-template")
	qt.Check(t, tmpl, qt.IsNil)

	tmpl = New("").Lookup("test-template")
	qt.Check(t, tmpl, qt.IsNil)

	tmpl = New("")
	tmpl, err := tmpl.New("test-template").Parse(`test-template-1`)
	qt.Assert(t, err, qt.IsNil)

	var b strings.Builder
	err = tmpl.Lookup("test-template").Execute(&b, nil)
	qt.Assert(t, err, qt.IsNil)
	qt.Check(t, b.String(), qt.Equals, "test-template-1")
}

func TestTemplateName(t *testing.T) {
	qt.Check(t, New("test-1").Name(), qt.Equals, "test-1")
	var tmpl Template
	qt.Check(t, tmpl.Name(), qt.Equals, "")
}

func TestTemplateNew(t *testing.T) {
	qt.Check(t, New("test-1").New("test-2").Name(), qt.Equals, "test-2")
}

func TestTemplateOption(t *testing.T) {
	tmpl, err := new(Template).Option("missingkey=error").Parse(`{{.key}}`)
	qt.Assert(t, err, qt.IsNil)

	var b strings.Builder
	err = tmpl.Execute(&b, map[string]string{})
	qt.Assert(t, err, qt.ErrorMatches, `template: :1:2: executing "" at <\.key>: map has no entry for key "key"`)
}

func TestTemplateParse(t *testing.T) {
	var tmpl Template
	_, err := tmpl.Parse(`{{.}}`)
	qt.Assert(t, err, qt.IsNil)
	qt.Check(t, tmpl.Name(), qt.Equals, "")

	var sb strings.Builder
	err = tmpl.Execute(&sb, "test-1")
	qt.Assert(t, err, qt.IsNil)
	qt.Check(t, sb.String(), qt.Equals, "'test-1'")

	_, err = tmpl.Parse(`{{`)
	qt.Check(t, err, qt.ErrorMatches, `template: :1: unclosed action`)
}

func TestTemplateParseFS(t *testing.T) {
	tmpl, err := New("test").ParseFS(testData, "testdata/*.tmpl")
	qt.Assert(t, err, qt.IsNil)
	qt.Check(t, tmpl.Name(), qt.Equals, "test")
	var names []string
	for _, tmpl := range tmpl.Templates() {
		names = append(names, tmpl.Name())
	}
	sort.Strings(names)
	qt.Check(t, names, qt.DeepEquals, []string{
		"B",
		"a.tmpl",
		"b.tmpl",
		"c.tmpl",
	})

	_, err = New("test").ParseFS(testData)
	qt.Check(t, err, qt.ErrorMatches, "template: no files named in call to ParseFiles")

	_, err = New("test").ParseFS(testData, "")
	qt.Check(t, err, qt.ErrorMatches, "template: pattern matches no files: ``")

	_, err = New("test").ParseFS(testData, "\\")
	qt.Check(t, err, qt.ErrorMatches, "syntax error in pattern")
}

func TestTemplateParseFiles(t *testing.T) {
	tmpl, err := New("test").ParseFiles("testdata/a.tmpl", "testdata/b.tmpl", "testdata/c.tmpl")
	qt.Assert(t, err, qt.IsNil)
	qt.Check(t, tmpl.Name(), qt.Equals, "test")
	var names []string
	for _, tmpl := range tmpl.Templates() {
		names = append(names, tmpl.Name())
	}
	sort.Strings(names)
	qt.Check(t, names, qt.DeepEquals, []string{
		"B",
		"a.tmpl",
		"b.tmpl",
		"c.tmpl",
	})

	_, err = New("test").ParseFiles()
	qt.Check(t, err, qt.ErrorMatches, "template: no files named in call to ParseFiles")
}

func TestTemplateParseGlob(t *testing.T) {
	tmpl, err := New("test").ParseGlob("testdata/*.tmpl")
	qt.Assert(t, err, qt.IsNil)
	qt.Check(t, tmpl.Name(), qt.Equals, "test")
	var names []string
	for _, tmpl := range tmpl.Templates() {
		names = append(names, tmpl.Name())
	}
	sort.Strings(names)
	qt.Check(t, names, qt.DeepEquals, []string{
		"B",
		"a.tmpl",
		"b.tmpl",
		"c.tmpl",
	})

	_, err = New("test").ParseGlob("")
	qt.Check(t, err, qt.ErrorMatches, "template: pattern matches no files: ``")

	_, err = New("test").ParseGlob("\\")
	qt.Check(t, err, qt.ErrorMatches, "syntax error in pattern")
}

func TestTemplateTemplates(t *testing.T) {
	tmpl, err := New("test").Parse(`{{.}}`)
	qt.Assert(t, err, qt.IsNil)

	tmpls := tmpl.Templates()
	qt.Check(t, tmpls, qt.HasLen, 1)

	_, err = tmpl.ParseFS(testData, "testdata/*.tmpl")
	qt.Assert(t, err, qt.IsNil)
	tmpls = tmpl.Templates()
	qt.Check(t, tmpls, qt.HasLen, 5)
}
