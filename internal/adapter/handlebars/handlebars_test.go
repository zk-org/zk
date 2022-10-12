package handlebars

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mickael-menu/zk/internal/adapter/handlebars/helpers"
	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/util"
	"github.com/mickael-menu/zk/internal/util/fixtures"
	"github.com/mickael-menu/zk/internal/util/paths"
	"github.com/mickael-menu/zk/internal/util/test/assert"
)

func init() {
	Init(true, &util.NullLogger)
}

// styler is a test double for core.Styler
// "hello", "red" -> "red(hello)"
type styler struct{}

func (s *styler) Style(text string, rules ...core.Style) (string, error) {
	return s.MustStyle(text, rules...), nil
}

func (s *styler) MustStyle(text string, rules ...core.Style) string {
	for _, rule := range rules {
		text = fmt.Sprintf("%s(%s)", rule, text)
	}
	return text
}

func testString(t *testing.T, template string, context interface{}, expected string) {
	sut := testLoader(LoaderOpts{})

	templ, err := sut.LoadTemplate(template)
	assert.Nil(t, err)

	actual, err := templ.Render(context)
	assert.Nil(t, err)
	assert.Equal(t, actual, expected)
}

func testFile(t *testing.T, name string, context interface{}, expected string) {
	sut := testLoader(LoaderOpts{})

	templ, err := sut.LoadTemplateAt(fixtures.Path(name))
	assert.Nil(t, err)

	actual, err := templ.Render(context)
	assert.Nil(t, err)
	assert.Equal(t, actual, expected)
}

func TestLookupPaths(t *testing.T) {
	root := fmt.Sprintf("/tmp/zk-test-%d", time.Now().Unix())
	os.Remove(root)
	path1 := filepath.Join(root, "1")
	os.MkdirAll(path1, os.ModePerm)
	path2 := filepath.Join(root, "1")
	os.MkdirAll(filepath.Join(path2, "subdir"), os.ModePerm)

	sut := testLoader(LoaderOpts{LookupPaths: []string{path1, path2}})

	test := func(path string, expected string) {
		tpl, err := sut.LoadTemplateAt(path)
		assert.Nil(t, err)
		res, err := tpl.Render(nil)
		assert.Nil(t, err)
		assert.Equal(t, res, expected)
	}

	test1 := filepath.Join(path1, "test1.tpl")

	tpl1, err := sut.LoadTemplateAt(test1)
	assert.Err(t, err, "cannot find template at "+test1)
	assert.Nil(t, tpl1)

	paths.WriteString(test1, "Test 1")
	test(test1, "Test 1")       // absolute
	test("test1.tpl", "Test 1") // relative

	test2 := filepath.Join(path2, "test2.tpl")
	paths.WriteString(test2, "Test 2")
	test(test2, "Test 2")       // absolute
	test("test2.tpl", "Test 2") // relative

	test3 := filepath.Join(path2, "subdir/test3.tpl")
	paths.WriteString(test3, "Test 3")
	test(test3, "Test 3")              // absolute
	test("subdir/test3.tpl", "Test 3") // relative
}

func TestRenderString(t *testing.T) {
	testString(t,
		"Goodbye, {{name}}",
		map[string]string{"name": "Ed"},
		"Goodbye, Ed",
	)
}

func TestRenderFile(t *testing.T) {
	testFile(t,
		"template.txt",
		map[string]string{"name": "Thom"},
		"Hello, Thom\n",
	)
}

func TestUnknownVariable(t *testing.T) {
	testString(t,
		"Hi, {{unknown}}!",
		nil,
		"Hi, !",
	)
}

func TestDoesntEscapeHTML(t *testing.T) {
	testString(t,
		"Salut, {{name}}!",
		map[string]string{"name": "l'ami"},
		"Salut, l'ami!",
	)

	testFile(t,
		"unescape.txt",
		map[string]string{"name": "l'ami"},
		"Salut, l'ami!\n",
	)
}

func TestConcatHelper(t *testing.T) {
	testString(t, "{{concat '> ' 'A quote'}}", nil, "> A quote")
}

func TestSubstringHelper(t *testing.T) {
	testString(t, "{{substring '' 2 4}}", nil, "")
	testString(t, "{{substring 'A full quote' 2 4}}", nil, "full")
	testString(t, "{{substring 'A full quote' 40 4}}", nil, "")
	testString(t, "{{substring 'A full quote' -5 5}}", nil, "quote")
	testString(t, "{{substring 'A full quote' -5 6}}", nil, "quote")
}

func TestJoinHelper(t *testing.T) {
	test := func(items []string, expected string) {
		context := map[string]interface{}{"items": items}
		testString(t, "{{join items '-'}}", context, expected)
	}

	test([]string{}, "")
	test([]string{"Item 1"}, "Item 1")
	test([]string{"Item 1", "Item 2"}, "Item 1-Item 2")
	test([]string{"Item 1", "Item 2", "Item 3"}, "Item 1-Item 2-Item 3")
}

type testJSONObject struct {
	Foo     string
	Missing string   `json:"missing,omitempty"`
	List    []string `json:"stringList"`
}

func TestJSONHelper(t *testing.T) {
	test := func(value interface{}, expected string) {
		context := map[string]interface{}{"value": value}
		testString(t, "{{json value}}", context, expected)
	}

	test(`foo"bar"`, `"foo\"bar\""`)
	test([]string{"foo", "bar"}, `["foo","bar"]`)
	test(map[string]string{"foo": "bar"}, `{"foo":"bar"}`)
	test(map[string]string{"foo": "bar"}, `{"foo":"bar"}`)
	test(testJSONObject{
		Foo:  "baz",
		List: []string{"foo", "bar"},
	}, `{"Foo":"baz","stringList":["foo","bar"]}`)
}

func TestPrependHelper(t *testing.T) {
	// inline
	testString(t, "{{prepend '> ' 'A quote'}}", nil, "> A quote")

	// block
	testString(t, "{{#prepend '> '}}A quote{{/prepend}}", nil, "> A quote")
	testString(t, "{{#prepend '> '}}A quote on\nseveral lines{{/prepend}}", nil, "> A quote on\n> several lines")
}

func TestListHelper(t *testing.T) {
	test := func(items []string, expected string) {
		context := map[string]interface{}{"items": items}
		testString(t, "{{list items}}", context, expected)
	}
	test([]string{}, "")
	test([]string{"Item 1"}, "  ‣ Item 1\n")
	test([]string{"Item 1", "Item 2"}, "  ‣ Item 1\n  ‣ Item 2\n")
	test([]string{"Item 1", "Item 2", "Item 3"}, "  ‣ Item 1\n  ‣ Item 2\n  ‣ Item 3\n")
	test([]string{"An item\non several\nlines\n"}, "  ‣ An item\n    on several\n    lines\n")
}

func TestLinkHelper(t *testing.T) {
	sut := testLoader(LoaderOpts{})

	templ, err := sut.LoadTemplate(`{{format-link "path/to note.md" "An interesting subject"}}`)
	assert.Nil(t, err)

	actual, err := templ.Render(map[string]interface{}{})
	assert.Nil(t, err)
	assert.Equal(t, actual, "path/to note.md - An interesting subject")
}

func TestSlugHelper(t *testing.T) {
	// inline
	testString(t,
		`{{slug "This will be slugified!"}}`,
		nil,
		"this-will-be-slugified",
	)
	// block
	testString(t,
		"{{#slug}}This will be slugified!{{/slug}}",
		nil,
		"this-will-be-slugified",
	)
}

func TestDateHelper(t *testing.T) {
	context := map[string]interface{}{"now": time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)}
	testString(t, "{{date now}}", context, "2009-11-17")
	testString(t, "{{date now 'short'}}", context, "11/17/2009")
	testString(t, "{{date now 'medium'}}", context, "Nov 17, 2009")
	testString(t, "{{date now 'long'}}", context, "November 17, 2009")
	testString(t, "{{date now 'full'}}", context, "Tuesday, November 17, 2009")
	testString(t, "{{date now 'year'}}", context, "2009")
	testString(t, "{{date now 'time'}}", context, "20:34")
	testString(t, "{{date now 'timestamp'}}", context, "200911172034")
	testString(t, "{{date now 'timestamp-unix'}}", context, "1258490098")
	testString(t, "{{date now 'cust: %Y-%m'}}", context, "cust: 2009-11")
	testString(t, "{{date now 'elapsed'}}", context, "13 years ago")
}

func TestGetDateHelper(t *testing.T) {
	context := map[string]interface{}{"now": time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)}
	testString(t, "{{get-date \"2009-11-17T20:34:58\"}}", context, "2009-11-17 20:34:58 +0000 UTC")
}

func TestShellHelper(t *testing.T) {
	// block is passed as piped input
	testString(t,
		`{{#sh "tr '[a-z]' '[A-Z]'"}}Hello, world!{{/sh}}`,
		nil,
		"HELLO, WORLD!",
	)
	// inline
	testString(t,
		`{{sh "echo 'Hello, world!'"}}`,
		nil,
		"Hello, world!",
	)

	// using pipes
	testString(t, `{{sh "echo hello | tr '[:lower:]' '[:upper:]'"}}`, nil, "HELLO")
}

func TestStyleHelper(t *testing.T) {
	// inline
	testString(t, "{{style 'single' 'Some text'}}", nil, "single(Some text)")
	testString(t, "{{style 'red bold' 'Another text'}}", nil, "bold(red(Another text))")

	// block
	testString(t, "{{#style 'single'}}A multiline\ntext{{/style}}", nil, "single(A multiline\ntext)")
}

func testLoader(opts LoaderOpts) *Loader {
	if opts.LookupPaths == nil {
		opts.LookupPaths = []string{}
	}
	if opts.Styler == nil {
		opts.Styler = &styler{}
	}

	loader := NewLoader(opts)

	loader.RegisterHelper("style", helpers.NewStyleHelper(opts.Styler, &util.NullLogger))
	loader.RegisterHelper("slug", helpers.NewSlugHelper("en", &util.NullLogger))

	formatter := func(context core.LinkFormatterContext) (string, error) {
		return context.Path + " - " + context.Title, nil
	}
	loader.RegisterHelper("format-link", helpers.NewLinkHelper(formatter, &util.NullLogger))

	return loader
}
