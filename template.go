package sqltemplate

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"text/template"
	"text/template/parse"
)

// A FuncMap is an alias for text/template.FuncMap. See
// https://golang.org/pkg/text/template#FuncMap for details.
type FuncMap = template.FuncMap

var funcs = FuncMap{
	"sqlliteral": PostgresLiteral,
}

// Must is a helper that wraps a call to a function returning (*Template, error)
// and panics if the error is non-nil. It is intended for use in variable
// initializations such as
//	var t = sqltemplate.Must(sqltemplate.New("name").Parse("text"))
func Must(t *Template, err error) *Template {
	if err != nil {
		panic(err)
	}
	return t
}

// ParseFS is like ParseFiles or ParseGlob but reads from the file system
// fsys instead of the host operating system's file system. It accepts a
// list of glob patterns. (Note that most file names serve as glob patterns
// matching only themselves.)
func ParseFS(fsys fs.FS, patterns ...string) (*Template, error) {
	if len(patterns) < 1 {
		return nil, fmt.Errorf("sqltemplate: no patterns provided in call to ParseFS")
	}
	filenames, err := fs.Glob(fsys, patterns[0])
	if err != nil {
		return nil, err
	}
	if len(filenames) < 1 {
		return nil, fmt.Errorf("sqltemplate: pattern matches no files: %#q", patterns[0])
	}
	return New(filepath.Base(filenames[0])).ParseFS(fsys, patterns...)
}

// ParseFiles creates a new Template and parses the template defintions
// from the named files. The returned template's name will have the base
// name and parsed contents of the first file. There must be at least one
// file. If an error occurs, parsing stops and the returned *Template is
// nil.
//
// When parsing multiple files with the same name in different directories,
// the last one mentioned will be the one that results. For instance,
// ParseFiles("a/foo", "b/foo") stores "b/foo" as the template named "foo",
// while "a/foo" is unavailable.
func ParseFiles(filenames ...string) (*Template, error) {
	if len(filenames) < 1 {
		return nil, fmt.Errorf("sqltemplate: no files named in call to ParseFiles")
	}
	return New(filepath.Base(filenames[0])).ParseFiles(filenames...)
}

// ParseGlob creates a new Template and parses the template definitions
// from the files identified by the pattern. The files are matched
// according to the semantics of filepath.Match, and the pattern must match
// at least one file. The returned template will have the (base) name and
// (parsed) contents of the first file matched by the pattern. ParseGlob is
// equivalent to calling ParseFiles with the list of files matched by the
// pattern.
//
// When parsing multiple files with the same name in different directories,
// the last one mentioned will be the one that results.
func ParseGlob(pattern string) (*Template, error) {
	filenames, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	if len(filenames) < 1 {
		return nil, fmt.Errorf("sqltemplate: pattern matches no files: %#q", pattern)
	}
	return New(filepath.Base(filenames[0])).ParseFiles(filenames...)
}

// A Template is the representation of a parsed template.
type Template struct {
	text *template.Template
}

func (t *Template) init() {
	if t.text == nil {
		t.text = new(template.Template).Funcs(funcs)
	}
}

// New allocates a new, undefined template with the given name.
func New(name string) *Template {
	return &Template{
		text: template.New(name).Funcs(funcs),
	}
}

// AddParseTree associates the argument parse tree with the template t,
// giving it the specified name. If the template has not been defined, this
// tree becomes its definition. If it has been defined and already has that
// name, the existing definition is replaced; otherwise a new template is
// created, defined, and returned.
func (t *Template) AddParseTree(name string, tree *parse.Tree) (*Template, error) {
	t.init()
	_, err := t.text.AddParseTree(name, escapeTree(tree.Copy()))
	return t, err
}

// Clone returns a duplicate of the template, including all associated
// templates. The actual representation is not copied, but the name space
// of associated templates is, so further calls to Parse in the copy will
// add templates to the copy but not to the original. Clone can be used to
// prepare common templates and use them with variant definitions for other
// templates by adding the variants after the clone is made.
func (t *Template) Clone() (*Template, error) {
	var t1 Template
	if t.text != nil {
		var err error
		t1.text, err = t.text.Clone()
		if err != nil {
			return nil, err
		}
	}
	return &t1, nil
}

// DefinedTemplates returns a string listing the defined templates,
// prefixed by the string "; defined templates are: ". If there are none,
// it returns the empty string. Used to generate an error message.
func (t *Template) DefinedTemplates() string {
	t.init()
	return t.text.DefinedTemplates()
}

// Delims sets the action delimiters to the specified strings, to be used
// in subsequent calls to Parse, ParseFiles, or ParseGlob. Nested template
// definitions will inherit the settings. An empty delimiter stands for the
// corresponding default: {{ or }}. The return value is the template, so
// calls can be chained.
func (t *Template) Delims(left, right string) *Template {
	t.init()
	t.text.Delims(left, right)
	return t
}

// Execute applies a parsed template to the specified data object, and
// writes the output to w. If an error occurs executing the template or
// writing its output, execution stops, but partial results may already
// have been written to the output writer. A template may be executed
// safely in parallel, although if parallel executions share a Writer the
// output may be interleaved.
//
// If data is a reflect.Value, the template applies to the concrete value
// that the reflect.Value holds, as in fmt.Print.
func (t *Template) Execute(w io.Writer, data interface{}) error {
	if t.text == nil {
		return fmt.Errorf("sqltemplate: %q is an incomplete or empty template", t.Name())
	}
	return t.text.Execute(w, data)
}

// ExecuteTemplate applies the template associated with t that has the
// given name to the specified data object and writes the output to w. If
// an error occurs executing the template or writing its output, execution
// stops, but partial results may already have been written to the output
// writer. A template may be executed safely in parallel, although if
// parallel executions share a Writer the output may be interleaved.
func (t *Template) ExecuteTemplate(w io.Writer, name string, data interface{}) error {
	tmpl := t.Lookup(name)
	if tmpl == nil {
		return fmt.Errorf("sqltemplate: no template %q associated with template %q", name, t.Name())
	}
	return tmpl.Execute(w, data)
}

// Funcs adds the elements of the argument map to the template's function
// map. It must be called before the template is parsed. It panics if a
// value in the map is not a function with appropriate return type or if
// the name cannot be used syntactically as a function in a template. It is
// legal to overwrite elements of the map. The return value is the
// template, so calls can be chained.
func (t *Template) Funcs(funcMap FuncMap) *Template {
	t.init()
	t.text.Funcs(funcMap)
	return t
}

// Lookup returns the template with the given name that is associated with
// t. It returns nil if there is no such template or the template has no
// definition.
func (t *Template) Lookup(name string) *Template {
	if t.text == nil {
		return nil
	}
	tt := t.text.Lookup(name)
	if tt == nil {
		return nil
	}
	return &Template{
		text: tt,
	}
}

// Name returns the name of the template.
func (t *Template) Name() string {
	if t.text == nil {
		return ""
	}
	return t.text.Name()
}

// New allocates a new, undefined template associated with the given one
// and with the same delimiters. The association, which is transitive,
// allows one template to invoke another with a {{template}} action.
//
// Because associated templates share underlying data, template
// construction cannot be done safely in parallel. Once the templates are
// constructed, they can be executed in parallel.
func (t *Template) New(name string) *Template {
	t.init()
	return &Template{
		text: t.text.New(name),
	}
}

// Option sets options for the template. Options are described by strings,
// either a simple string or "key=value". There can be at most one equals
// sign in an option string. If the option string is unrecognized or
// otherwise invalid, Option panics.
//
// This package does not define any options, the only options supported are
// those listed in https://golang.org/pkg/text/template#Template.Option.
func (t *Template) Option(opt ...string) *Template {
	t.init()
	t.text.Option(opt...)
	return t
}

// Parse parses text as a template body for t. Named template definitions
// ({{define ...}} or {{block ...}} statements) in text define additional
// templates associated with t and are removed from the definition of t
// itself.
//
// Templates can be redefined in successive calls to Parse. A template
// definition with a body containing only white space and comments is
// considered empty and will not replace an existing template's body. This
// allows using Parse to add new named template definitions without
// overwriting the main template body.
func (t *Template) Parse(text string) (*Template, error) {
	t.init()
	tt, err := t.text.Parse(text)
	if err != nil {
		return nil, err
	}
	escapeTemplate(tt)
	return t, nil
}

// ParseFS is like ParseFiles or ParseGlob but reads from the file system
// fsys instead of the host operating system's file system. It accepts a
// list of glob patterns. (Note that most file names serve as glob patterns
// matching only themselves.)
func (t *Template) ParseFS(fsys fs.FS, patterns ...string) (*Template, error) {
	t.init()
	tt, err := t.text.ParseFS(fsys, patterns...)
	if err != nil {
		return nil, err
	}
	escapeTemplate(tt)
	return t, nil
}

// ParseFiles parses the named files and associates the resulting templates
// with t. If an error occurs, parsing stops and the returned template is
// nil; otherwise it is t. There must be at least one file. Since the
// templates created by ParseFiles are named by the base names of the
// argument files, t should usually have the name of one of the (base)
// names of the files. If it does not, depending on t's contents before
// calling ParseFiles, t.Execute may fail. In that case use
// t.ExecuteTemplate to execute a valid template.
//
// When parsing multiple files with the same name in different directories,
// the last one mentioned will be the one that results.
func (t *Template) ParseFiles(filenames ...string) (*Template, error) {
	t.init()
	tt, err := t.text.ParseFiles(filenames...)
	if err != nil {
		return nil, err
	}
	escapeTemplate(tt)
	return t, nil
}

// ParseGlob parses the template definitions in the files identified by the
// pattern and associates the resulting templates with t. The files are
// matched according to the semantics of filepath.Match, and the pattern
// must match at least one file. ParseGlob is equivalent to calling
// t.ParseFiles with the list of files matched by the pattern.
//
// When parsing multiple files with the same name in different directories,
// the last one mentioned will be the one that results.
func (t *Template) ParseGlob(pattern string) (*Template, error) {
	t.init()
	tt, err := t.text.ParseGlob(pattern)
	if err != nil {
		return nil, err
	}
	escapeTemplate(tt)
	return t, nil
}

// Templates returns a slice of defined templates associated with t.
func (t *Template) Templates() []*Template {
	t.init()
	tts := t.text.Templates()
	ts := make([]*Template, len(tts))
	for i, tt := range tts {
		ts[i] = &Template{
			text: tt,
		}
	}
	return ts
}

// escapeTemplate escapes all the templates defined in a template.
func escapeTemplate(t *template.Template) {
	for _, tmpl := range t.Templates() {
		escapeTree(tmpl.Tree)
	}
}
