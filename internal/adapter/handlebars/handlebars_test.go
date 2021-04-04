package handlebars

import (
	"fmt"
	"testing"
	"time"

	"github.com/mickael-menu/zk/internal/core/style"
	"github.com/mickael-menu/zk/internal/util"
	"github.com/mickael-menu/zk/internal/util/fixtures"
	"github.com/mickael-menu/zk/internal/util/test/assert"
)

func init() {
	Init("en", true, &util.NullLogger, &styler{})
}

// styler is a test double for core.Styler
// "hello", "red" -> "red(hello)"
type styler struct{}

func (s *styler) Style(text string, rules ...style.Rule) (string, error) {
	return s.MustStyle(text, rules...), nil
}

func (s *styler) MustStyle(text string, rules ...style.Rule) string {
	for _, rule := range rules {
		text = fmt.Sprintf("%s(%s)", rule, text)
	}
	return text
}

func testString(t *testing.T, template string, context interface{}, expected string) {
	sut := NewLoader()

	templ, err := sut.Load(template)
	assert.Nil(t, err)

	actual, err := templ.Render(context)
	assert.Nil(t, err)
	assert.Equal(t, actual, expected)
}

func testFile(t *testing.T, name string, context interface{}, expected string) {
	sut := NewLoader()

	templ, err := sut.LoadFile(fixtures.Path(name))
	assert.Nil(t, err)

	actual, err := templ.Render(context)
	assert.Nil(t, err)
	assert.Equal(t, actual, expected)
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
	testString(t, "{{date now 'elapsed'}}", context, "12 years ago")
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
